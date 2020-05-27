package customconfig

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
)

// Provider returns the actual provider instance.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"timeout_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60, // 1 + (n*2) roof 16 = 1+2+4+8+16 = 31 seconds, 1 min should be "normal" operations
			},
		},
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"customconfig_google_backend":           dataSourceGoogleBackend(),
			"customconfig_google_forwarding_config": dataSourceGoogleBackend(),
			"customconfig_vault_token":              vaultTokenDataSource(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	timeoutSeconds := d.Get("timeout_seconds").(int)

	config := Config{
		TimeoutSeconds: timeoutSeconds,
	}

	if err := config.loadAndValidate(); err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	return &config, nil
}
