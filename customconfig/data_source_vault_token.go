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
		},
	}
}

func vaultTokenDataSourceRead(d *schema.ResourceData, meta interface{}) error {
	roleID := d.Get("role_id").(string)
	secretID := d.Get("secret_id").(string)
	log.Printf("[DEBUG] Reading %s %d from Vault", roleID, secretID)

	backend := d.Get("backend").(string)

	requestBody, err := json.Marshal(map[string]string{
		"role_id":   roleID,
		"secret_id": secretID,
	})
	if err != nil {
		return fmt.Errorf("error creating request body json: %s", err)
	}

	addr := os.Getenv("VAULT_ADDR")
	resp, err := http.Post(fmt.Sprintf("%s/v1/auth/%s/login", addr, backend), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error talking to Vault: %s", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading from Vault: %s", err)
	}

	data := string(body)
	d.Set("data_json", string(data))

	var secret map[string]interface{}
	json.Unmarshal([]byte(data), &secret)
	if secret["errors"] != nil {
		errors := secret["errors"].(map[string]interface{})
		if len(errors) > 0 {
			return fmt.Errorf("vault returned error(s): %v", errors)
		}
	}

	d.SetId(secret["request_id"].(string))
	d.Set("lease_duration", secret["lease_duration"].(float64))
	d.Set("lease_renewable", secret["renewable"].(bool))
	d.Set("lease_start_time", time.Now().Format("RFC3339"))

	auth := secret["auth"].(map[string]interface{})

	d.Set("token", auth["client_token"].(string))

	return nil
}
