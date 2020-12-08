package customconfig

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func vaultTokenDataSource() *schema.Resource {
	return &schema.Resource{
		Read: vaultTokenDataSourceRead,

		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_ADDR", nil),
				Description: "URL of the root of the target Vault server.",
			},

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

			"debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Flag to write more data to the debug log",
			},

			"ca_cert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_CACERT", ""),
				Description: "Path to a CA certificate file to validate the server's certificate.",
			},

			"ca_cert_dir": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_CAPATH", ""),
				Description: "Path to directory containing CA certificate files to validate the server's certificate.",
			},

			"skip_tls_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_SKIP_VERIFY", false),
				Description: "Set this to true only if the target Vault server is an insecure development instance.",
			},

			"data_json": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JSON-encoded secret data read from Vault.",
				Sensitive:   true,
			},

			"token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role/secret generated Vault auth token.",
				Sensitive:   true,
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
	config := meta.(*Config)

	roleID := d.Get("role_id").(string)
	secretID := d.Get("secret_id").(string)

	debug := d.Get("debug").(bool)
	if debug {
		log.Printf("[DEBUG] Reading %s %d from Vault", roleID, secretID)
	}

	// handle SSL
	skipTLSVerify := d.Get("skip_tls_verify").(bool)
	CACertDir := d.Get("ca_cert_dir").(string)
	CACertFile := d.Get("ca_cert_file").(string)
	RootCAPath := filepath.Join(CACertDir, CACertFile)

	if RootCAPath != "" {
		crt, err := ioutil.ReadFile(RootCAPath)
		if err != nil {
			log.Fatal(err)
		}

		rootCAs := x509.NewCertPool()
		rootCAs.AppendCertsFromPEM(crt)
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: skipTLSVerify,
			RootCAs:            rootCAs,
		}
	} else {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: skipTLSVerify}
	}

	backend := d.Get("backend").(string)

	requestBody, err := json.Marshal(map[string]string{
		"role_id":   roleID,
		"secret_id": secretID,
	})
	if err != nil {
		return fmt.Errorf("error creating request body json: %s", err)
	}

	addr := d.Get("address").(string)
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
