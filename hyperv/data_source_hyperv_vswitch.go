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
                ForceNew: true,
            },
            "switch_type": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,

                StateFunc:        tfutil.StateToLower(),
            },
            "notes": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "allow_management_os": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            "net_adapter_name": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "net_adapter_interface_description": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
        },
    }
}

func dataSourceHypervVSwitchRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    id                            := fmt.Sprintf("//%s/vswitches/%s", host, d.Get("name").(string))
    name := d.Get("name").(string)

    log.Printf("[INFO][terraform-provider-hyperv] reading hyperv_vswitch %q\n", id)

    // read vswitch
    vs := new(api.VSwitch)
    vs.Name = name

    vswitch, err := c.ReadVSwitch(vs)
    if err != nil {
        log.Printf("[INFO][terraform-provider-hyperv] cannot read hyperv_vswitch %q\n", id)
        return err
    }

    // set properties
    d.Set("name", vswitch.Name)
    d.Set("switch_type", strings.ToLower(vswitch.SwitchType))
    d.Set("notes", vswitch.Notes)
    d.Set("allow_management_os", vswitch.AllowManagementOS)
    d.Set("net_adapter_name", vswitch.NetAdapterName)
    d.Set("net_adapter_interface_description", vswitch.NetAdapterInterfaceDescription)

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-hyperv] read hyperv_vswitch %q\n", id)
    return nil
}

//------------------------------------------------------------------------------
