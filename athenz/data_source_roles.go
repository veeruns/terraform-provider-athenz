package athenz

import (
	"fmt"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceRoles() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRolesRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tag_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tag_value": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"include_members": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"members": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"tags": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceRolesRead(d *schema.ResourceData, meta interface{}) error {
	zmsClient := meta.(client.ZmsClient)

	dn := d.Get("domain").(string)
	tagKey := d.Get("tag_key").(string)
	tagValue := d.Get("tag_value").(string)
	if tagValue != "" && tagKey == "" {
		return fmt.Errorf("in order to input tag_value, tag_key must be provided")
	}
	members := d.Get("include_members").(bool)
	roles, err := zmsClient.GetRoles(dn, &members, tagKey, tagValue)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			return fmt.Errorf("athenz Roles %s not found, update your data source query", dn+"key: "+tagKey+", value: "+tagValue)
		} else {
			return fmt.Errorf("error retrieving Athenz Role: %s", v)
		}
	case rdl.Any:
		return err
	}
	fullResourceName := dn + "_" + tagKey + "_" + tagValue
	d.SetId(fullResourceName)
	if roles != nil && roles.List != nil {
		if err = d.Set("roles", flattenRoles(roles.List, dn)); err != nil {
			return err
		}
	}
	return nil
}
