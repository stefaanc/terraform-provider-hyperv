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

func dataSourceHypervVNetworkAdapter() *schema.Resource {
    return &schema.Resource{
        Read:   dataSourceHypervVNetworkAdapterRead,

        Schema: map[string]*schema.Schema{
            "name": &schema.Schema{
                Type:     schema.TypeString,   // remark that there can be multiple vnetwork_adapters with the same name (for different vmachines/management_os)
                Optional: true,                // "name" and "vmachine_name" needs to identify exactly one vnetwork_adapter
                Computed: true,
            },
            "vmachine_name": &schema.Schema{   // use "" for management_os
                Type:     schema.TypeString,   // remark that a vmachine/management_os can have multiple vnetwork_adapters (with different names)
                Optional: true,                // "name" and "vmachine_name" needs to identify exactly one vnetwork_adapter
                Computed: true,
            },

            "mac_address": &schema.Schema{   // when dynamic (as opposed to static ), has a value only when connected to a vswitch
                Type:     schema.TypeString,
                Computed: true,
            },
            "allow_mac_address_spoofing": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },

            "vswitch_name": &schema.Schema{   // has a value only when connected to a vswitch
                Type:     schema.TypeString,
                Computed: true,
            },

            // lifecycle customizations that are not supported by the 'lifecycle' meta-argument for data sources
            "x_lifecycle": &tfutil.DataSourceXLifecycleSchema,
        },
    }
}

//------------------------------------------------------------------------------

func dataSourceHypervVNetworkAdapterRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    name          := d.Get("name").(string)
    vmachine_name := d.Get("vmachine_name").(string)
    x_lifecycle   := tfutil.GetResourceDataMap(d, "x_lifecycle")

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    full_name := name
    if vmachine_name != "" {
        full_name += "@" + vmachine_name
    }

    id := fmt.Sprintf("//%s/vnetwork_adapters/%s", host, full_name)

    log.Printf("[INFO][terraform-provider-hyperv] reading hyperv_vnetwork_adapter %q\n", id)

    // read vnetwork_adapter
    vnaQuery := new(api.VNetworkAdapter)
    vnaQuery.Name = name
    vnaQuery.VMachineName = vmachine_name

    vnetworkAdapter, err := c.ReadVNetworkAdapter(vnaQuery)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        if x_lifecycle != nil {
            ignore_error_if_not_exists := x_lifecycle["ignore_error_if_not_exists"].(bool)
            if ignore_error_if_not_exists && strings.Contains(err.Error(), "cannot find vnetwork_adapter") {
                log.Printf("[INFO][terraform-provider-hyperv] cannot read hyperv_vnetwork_adapter %q\n", id)

                // set zeroed properties
                d.Set("name", "")
                d.Set("vmachine_name", "")

                d.Set("mac_address", "")
                d.Set("allow_mac_address_spoofing", false)

                d.Set("vswitch_name", "")

                // set computed lifecycle properties
                x_lifecycle["exists"] = false
                tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)

                // set id
                d.SetId(id)

                log.Printf("[INFO][terraform-provider-hyperv] ignored error and added zeroed hyperv_vnetwork_adapter %q to terraform state\n", id)
                return nil
            }
        }
        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-hyperv] cannot read hyperv_vnetwork_adapter %q\n", id)
        return err
    }

    // set properties
    d.Set("name", vnetworkAdapter.Name)
    d.Set("vmachine_name", vnetworkAdapter.VMachineName)

    d.Set("mac_address", vnetworkAdapter.MACAddress)
    d.Set("allow_mac_address_spoofing", vnetworkAdapter.AllowMACAddressSpoofing)

    d.Set("vswitch_name", vnetworkAdapter.VSwitchName)

    // set computed lifecycle properties
    if x_lifecycle != nil {
        x_lifecycle["exists"] = true
        tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)
    }

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-hyperv] read hyperv_vnetwork_adapter %q\n", id)
    return nil
}

//------------------------------------------------------------------------------
