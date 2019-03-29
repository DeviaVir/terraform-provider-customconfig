package customconfig

import (
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceGoogleForwardingConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGoogleForwardingConfigRead,
		Schema: map[string]*schema.Schema{
			"ipv4_addresses": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"target_name_servers": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
				Computed: true,
			},
		},
	}
}

func dataSourceGoogleForwardingConfigRead(d *schema.ResourceData, meta interface{}) error {
	ips := d.Get("ipv4_addresses").([]interface{})
	ipsStr := make([]string, len(ips))
	outputs := make([]map[string]string, len(ips))
	for idx, ip := range ips {
		outputs[idx] = map[string]string{}
		outputs[idx]["ipv4_address"] = ip.(string)
		ipsStr[idx] = ip.(string)
	}

	d.SetId(hash(strings.Join(ipsStr, ",")))
	d.Set("target_name_servers", outputs)
	return nil
}
