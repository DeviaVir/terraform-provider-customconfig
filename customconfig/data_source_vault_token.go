package customconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func vaultTokenDataSource() *schema.Resource {
	return &schema.Resource{
		Read: vaultTokenDataSourceRead,

		Schema: map[string]*schema.Schema{
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role ID of the Vault Approle.",
			},

			"secret_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Secret ID of the Approle's auth backend secret.",
			},

			"backend": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "ptfe",
				Description: "Approle Backend to use, ptfe by default.",
			},

			"data_json": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JSON-encoded secret data read from Vault.",
			},

			"token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role/secret generated Vault auth token.",
			},

			"renewable": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "Flag to allow the token to be renewed",
			},

			"lease_duration": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The token lease duration.",
			},

			"lease_start_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The token lease started on.",
			},

			"debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Flag to write more data to the debug log",
			},
		},
	}
}

func vaultTokenDataSourceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	roleID := d.Get("role_id").(string)
	secretID := d.Get("secret_id").(string)

	debug := d.Get("debug").(bool)
	if debug {
		log.Printf("[DEBUG] Reading %s %d from Vault", roleID, secretID)
	}

	backend := d.Get("backend").(string)

	requestBody, err := json.Marshal(map[string]string{
		"role_id":   roleID,
		"secret_id": secretID,
	})
	if err != nil {
		return fmt.Errorf("error creating request body json: %s", err)
	}

	addr := os.Getenv("VAULT_ADDR")
	err = retry(func() error {
		dataJSON, secret, err := vaultTokenDataSourceReadCall(requestBody, addr, backend, debug)
		if debug {
			log.Printf("[DEBUG] Received dataJSON %s", dataJSON)
			log.Printf("[DEBUG] Received secret %+v", secret)
		}

		if secret != nil && dataJSON != "" {
			d.SetId(secret["request_id"].(string))
			d.Set("data_json", dataJSON)
			d.Set("lease_duration", secret["lease_duration"].(float64))
			d.Set("lease_renewable", secret["renewable"].(bool))
			d.Set("lease_start_time", time.Now().Format("RFC3339"))

			auth := secret["auth"].(map[string]interface{})

			d.Set("token", auth["client_token"].(string))
		}

		return err
	}, config.TimeoutSeconds)
	if err != nil {
		return err
	}

	return nil
}

func vaultTokenDataSourceReadCall(requestBody []byte, addr, backend string, debug bool) (string, map[string]interface{}, error) {
	var secret map[string]interface{}
	resp, err := http.Post(fmt.Sprintf("%s/v1/auth/%s/login", addr, backend), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", secret, fmt.Errorf("error talking to Vault: %s", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", secret, fmt.Errorf("error reading from Vault: %s", err)
	}

	data := string(body)
	if debug {
		log.Printf("[DEBUG] Got %s from Vault", string(data))
	}

	json.Unmarshal(body, &secret)
	if secret["errors"] != nil {
		return "", secret, fmt.Errorf("vault returned error(s): %#v", secret["errors"])
	}

	return string(data), secret, nil
}
