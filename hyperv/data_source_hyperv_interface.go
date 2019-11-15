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
    "strconv"
    "strings"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

    "github.com/stefaanc/terraform-provider-hyperv/api"
    "github.com/stefaanc/terraform-provider-hyperv/hyperv/tfutil"
)

//------------------------------------------------------------------------------

func dataSourceHypervInterface() *schema.Resource {
    return &schema.Resource{
        Read:   dataSourceHypervInterfaceRead,

        Schema: map[string]*schema.Schema{
            "index": &schema.Schema{
                Type:     schema.TypeInt,
                Optional: true,
                Computed: true,
            },
            "name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index" },
            },
            "alias": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index", "name" },
            },
            "description": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index", "name", "alias" },
            },

            "mac_address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index", "name", "alias", "description" },
            },
            "network_adapter_name": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index", "name", "alias", "description", "mac_address" },
            },
            "vnetwork_adapter_name": &schema.Schema{   // only for a net_adapter of the management_os
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "index", "name", "alias", "description", "mac_address","network_adapter_name" },
            },

            "network_name": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "computer_name": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            // lifecycle customizations that are not supported by the 'lifecycle' meta-argument for data sources
            "x_lifecycle": &tfutil.DataSourceXLifecycleSchema,
        },
    }
}

//------------------------------------------------------------------------------

func dataSourceHypervInterfaceRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    index               := uint32(d.Get("index").(int))
    name                := d.Get("name").(string)
    alias               := d.Get("alias").(string)
    description         := d.Get("description").(string)
    macAddress          := d.Get("mac_address").(string)
    networkAdapterName  := d.Get("network_adapter_name").(string)
    vnetworkAdapterName := d.Get("vnetwork_adapter_name").(string)
    x_lifecycle         := tfutil.GetResourceDataMap(d, "x_lifecycle")

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    var id string
    if        index != 0                { id = strconv.FormatUint(uint64(index), 10)
    } else if name != ""                { id = name
    } else if alias != ""               { id = alias
    } else if description != ""         { id = description
    } else if macAddress != ""          { id = macAddress
    } else if networkAdapterName != ""  { id = networkAdapterName
    } else if vnetworkAdapterName != "" { id = vnetworkAdapterName
    }
    id = fmt.Sprintf("//%s/interfaces/%s", host, id)

    log.Printf("[INFO][terraform-provider-hyperv] reading hyperv_interface %q\n", id)

    // read interface
    iQuery := new(api.Interface)
    iQuery.Index               = index
    iQuery.Name                = name
    iQuery.Alias               = alias
    iQuery.Description         = description
    iQuery.MACAddress          = macAddress
    iQuery.NetworkAdapterName  = networkAdapterName
    iQuery.VNetworkAdapterName = vnetworkAdapterName

    iProperties, err := c.ReadInterface(iQuery)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        if x_lifecycle != nil {
            ignore_error_if_not_exists := x_lifecycle["ignore_error_if_not_exists"].(bool)
            if ignore_error_if_not_exists && strings.Contains(err.Error(), "cannot find interface") {
                log.Printf("[INFO][terraform-provider-hyperv] cannot read hyperv_interface %q\n", id)

                // set zeroed properties
                d.Set("index", 0)
                d.Set("name", "")
                d.Set("alias", "")
                d.Set("description", "")
                d.Set("mac_address", "")
                d.Set("network_adapter_name", "")
                d.Set("vnetwork_adapter_name", "")
                d.Set("network_name", "")
                d.Set("computer_name", "")

                // set computed lifecycle properties
                x_lifecycle["exists"] = false
                tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)

                // set id
                d.SetId(id)

                log.Printf("[INFO][terraform-provider-hyperv] ignored error and added zeroed hyperv_interface %q to terraform state\n", id)
                return nil
            }
        }

        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-hyperv] cannot read hyperv_interface %q\n", id)
        return err
    }

    // set properties
    d.Set("index", iProperties.Index)
    d.Set("name", iProperties.Name)
    d.Set("alias", iProperties.Alias)
    d.Set("description", iProperties.Description)
    d.Set("mac_address", iProperties.MACAddress)
    d.Set("network_adapter_name", iProperties.NetworkAdapterName)
    d.Set("vnetwork_adapter_name", iProperties.VNetworkAdapterName)
    d.Set("network_name", iProperties.NetworkName)
    d.Set("computer_name", iProperties.ComputerName)

    // set computed lifecycle properties
    if x_lifecycle != nil {
        x_lifecycle["exists"] = true
        tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)
    }

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-hyperv] read hyperv_interface %q\n", id)
    return nil
}

//------------------------------------------------------------------------------
