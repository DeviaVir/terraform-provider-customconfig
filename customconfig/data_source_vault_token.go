package customconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/vault/api"
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

			"data": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Map of strings read from Vault.",
			},
		},
	}
}

func vaultTokenDataSourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	roleID := d.Get("role_id").(string)
	secretID := d.Get("secret_id").(string)
	log.Printf("[DEBUG] Reading %s %d from Vault", roleID, secretID)

	backend := d.Get("backend").(string)

	authParams := map[string]string{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	r := client.NewRequest("GET", "/v1/"+fmt.Sprintf("auth/%s/login", backend))
	for k, v := range authParams {
		r.Params.Set(k, v)
	}
	resp, err := client.RawRequest(r)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("error reading from Vault: %s", err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	data := buf.String()
	d.Set("data_json", string(data))

	var secret map[string]interface{}
	json.Unmarshal([]byte(data), &secret)
	errors := secret["errors"].(map[string]interface{})
	if len(errors) > 0 {
		return fmt.Errorf("vault returned error(s): %v", errors)
	}

	d.SetId(secret["request_id"].(string))
	d.Set("lease_id", secret["lease_id"].(string))
	d.Set("lease_duration", secret["lease_duration"].(string))
	d.Set("lease_renewable", secret["lease_renewable"].(string))
	d.Set("lease_start_time", time.Now().Format("RFC3339"))

	auth := secret["auth"].(map[string]interface{})

	d.Set("token", auth["client_token"].(string))

	return nil
}
