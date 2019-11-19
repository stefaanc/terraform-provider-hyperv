//
// Copyright (c) 2019 Sean Reynolds, Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-hyperv
//
package hyperv

import (
    "fmt"
    "log"
    "strings"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

    "github.com/stefaanc/terraform-provider-hyperv/api"
    "github.com/stefaanc/terraform-provider-hyperv/hyperv/tfutil"
)

//------------------------------------------------------------------------------

func dataSourceHypervVSwitch () *schema.Resource {
    return &schema.Resource{
        Read:   dataSourceHypervVSwitchRead,

        Schema: map[string]*schema.Schema{
            "name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
            },

            "network_adapter_name": &schema.Schema{
                Type:     schema.TypeSet,
                Computed: true,
                Elem: &schema.Schema{ Type: schema.TypeString },
            },

            "notes": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            // lifecycle customizations that are not supported by the 'lifecycle' meta-argument for data sources
            "x_lifecycle": &tfutil.DataSourceXLifecycleSchema,
        },
    }
}

//------------------------------------------------------------------------------

func dataSourceHypervVSwitchRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    name        := d.Get("name").(string)
    x_lifecycle := tfutil.GetResourceDataMap(d, "x_lifecycle")

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    id          := fmt.Sprintf("//%s/vswitches/%s", host, name)

    log.Printf("[INFO][terraform-provider-hyperv] reading hyperv_vswitch %q\n", id)

    // read vswitch
    vs := new(api.VSwitch)
    vs.Name = name

    vswitch, err := c.ReadVSwitch(vs)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        if x_lifecycle != nil {
            ignore_error_if_not_exists := x_lifecycle["ignore_error_if_not_exists"].(bool)
            if ignore_error_if_not_exists && strings.Contains(err.Error(), "cannot find vswitch") {
                log.Printf("[INFO][terraform-provider-hyperv] cannot read hyperv_vswitch %q\n", id)

                // set zeroed properties
                d.Set("name", "")
                d.Set("network_adapter_name", schema.NewSet(schema.HashString, nil))
                d.Set("notes", "")

                // set computed lifecycle properties
                x_lifecycle["exists"] = false
                tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)

                // set id
                d.SetId(id)

                log.Printf("[INFO][terraform-provider-hyperv] ignored error and added zeroed hyperv_vswitch %q to terraform state\n", id)
                return nil
            }
        }

        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-hyperv] cannot read hyperv_vswitch %q\n", id)
        return err
    }

    // set properties
    d.Set("name", vswitch.Name)

    networkAdapters := make([]interface{}, len(vswitch.NetworkAdapters))
    for i, n := range vswitch.NetworkAdapters {
        networkAdapters[i] = n
    }
    d.Set("network_adapter_name", schema.NewSet(schema.HashString, networkAdapters))

    d.Set("notes", vswitch.Notes)

    // set computed lifecycle properties
    if x_lifecycle != nil {
        x_lifecycle["exists"] = true
        tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)
    }

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-hyperv] read hyperv_vswitch %q\n", id)
    return nil
}

//------------------------------------------------------------------------------
