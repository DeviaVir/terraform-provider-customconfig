package customconfig

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
)

var (
	// contextTimeout is the global context timeout for requests to complete.
	contextTimeout = 15 * time.Second
)

// Provider returns the actual provider instance.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"timeout_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1, // 1 + (n*2) roof 16 = 1+2+4+8+16 = 31 seconds, 1 min should be "normal" operations
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
	timeoutMinutes := d.Get("timeout_minutes").(int)

	config := Config{
		TimeoutMinutes: timeoutMinutes,
	}

	if err := config.loadAndValidate(); err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	return &config, nil
}

// contextWithTimeout creates a new context with the global context timeout.
func contextWithTimeout() (context.Context, func()) {
	return context.WithTimeout(context.Background(), contextTimeout)
}
