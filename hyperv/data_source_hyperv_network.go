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

func dataSourceHypervNetwork() *schema.Resource {
    return &schema.Resource{
        Read:   dataSourceHypervNetworkRead,

        Schema: map[string]*schema.Schema{
            "name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
            },

            "connection_profile": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            // lifecycle customizations that are not supported by the 'lifecycle' meta-argument for data sources
            "x_lifecycle": &tfutil.DataSourceXLifecycleSchema,
        },
    }
}

//------------------------------------------------------------------------------

func dataSourceHypervNetworkRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    name        := d.Get("name").(string)
    x_lifecycle := tfutil.GetResourceDataMap(d, "x_lifecycle")

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    id := fmt.Sprintf("//%s/networks/%s", host, name)

    log.Printf("[INFO][terraform-provider-hyperv] reading hyperv_network %q\n", id)

    // read network
    nQuery := new(api.Network)
    nQuery.Name = name

    network, err := c.ReadNetwork(nQuery)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        if x_lifecycle != nil {
            ignore_error_if_not_exists := x_lifecycle["ignore_error_if_not_exists"].(bool)
            if ignore_error_if_not_exists && strings.Contains(err.Error(), "cannot find network") {
                log.Printf("[INFO][terraform-provider-hyperv] cannot read hyperv_network %q\n", id)

                // set zeroed properties
                d.Set("name", "")
                d.Set("connection_profile", "")

                // set computed lifecycle properties
                x_lifecycle["exists"] = false
                tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)

                // set id
                d.SetId(id)

                log.Printf("[INFO][terraform-provider-hyperv] ignored error and added zeroed hyperv_network %q to terraform state\n", id)
                return nil
            }
        }
        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-hyperv] cannot read hyperv_network %q\n", id)
        return err
    }

    // set properties
    d.Set("name", network.Name)
    d.Set("connection_profile", network.ConnectionProfile)

    // set computed lifecycle properties
    if x_lifecycle != nil {
        x_lifecycle["exists"] = true
        tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)
    }

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-hyperv] read  hyperv_network %q\n", id)
    return nil
}

//------------------------------------------------------------------------------
