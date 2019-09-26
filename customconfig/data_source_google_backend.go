package customconfig

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceGoogleBackend() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGoogleBackendRead,
		Schema: map[string]*schema.Schema{
			"instance_groups": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"backends": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
				Computed: true,
			},
		},
	}
}

func dataSourceGoogleBackendRead(d *schema.ResourceData, meta interface{}) error {
	groups := d.Get("instance_groups").([]interface{})
	groupsStr := make([]string, len(groups))
	outputs := make([]map[string]string, len(groups))
	for idx, group := range groups {
		outputs[idx] = map[string]string{}
		outputs[idx]["group"] = group.(string)
		groupsStr[idx] = group.(string)
	}

	d.SetId(hash(strings.Join(groupsStr, ",")))
	d.Set("backends", outputs)
	return nil
}
