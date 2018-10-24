package customconfig

import (
	"context"
	"time"

	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
	"log"
	"os"
)

var (
	// contextTimeout is the global context timeout for requests to complete.
	contextTimeout = 15 * time.Second
)

// Provider returns the actual provider instance.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{},
    DataSourceMap: map[string]*schema.Resource{
      "customconfig_google_backend": dataSourceGoogleBackend(),
    },
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{}

	if err := config.loadAndValidate(); err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	return &config, nil
}

// contextWithTimeout creates a new context with the global context timeout.
func contextWithTimeout() (context.Context, func()) {
	return context.WithTimeout(context.Background(), contextTimeout)
}
