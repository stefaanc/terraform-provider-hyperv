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
    "github.com/hashicorp/terraform-plugin-sdk/helper/validation"

    "github.com/stefaanc/terraform-provider-hyperv/api"
    "github.com/stefaanc/terraform-provider-hyperv/hyperv/tfutil"
)

//------------------------------------------------------------------------------

func resourceHypervVSwitch () *schema.Resource {
    return &schema.Resource{
        Create: resourceHypervVSwitchCreate,
        Read:   resourceHypervVSwitchRead,
        Update: resourceHypervVSwitchUpdate,
        Delete: resourceHypervVSwitchDelete,

        Importer: &schema.ResourceImporter{
            State: resourceHypervVSwitchImport,
        },

        Schema: map[string]*schema.Schema{
            "name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,

                DiffSuppressFunc: tfutil.DiffSuppressCase(),
            },
            "switch_type": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Default:  "internal",
                ForceNew: true,                                    // not a problem to update the switch when changing from "external" to "internal"
                                                                   // but need to delete and re-create the switch when changing from "internal" to "external"

                ValidateFunc:     validation.StringInSlice([]string{ "private", "internal", "external" }, true),
                StateFunc:        tfutil.StateToLower(),
                DiffSuppressFunc: tfutil.DiffSuppressCase(),
            },
            "notes": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                Default:  "",
            },
            "allow_management_os": &schema.Schema{                 // config ignored when switch_type is not "external"
                Type:     schema.TypeBool,                         // by default set to true when switch_type is "internal", to false when switch_type is "private"
                Optional: true,
                Computed: true,
            },
            "net_adapter_name": &schema.Schema{                    // config ignored when switch_type is not "external"
                Type:     schema.TypeString,                       // when set: uses existing adapter
                Optional: true,
                Computed: true,

                DiffSuppressFunc: tfutil.DiffSuppressCase(),
            },
            "net_adapter_interface_description": &schema.Schema{   // config ignored when switch_type is not "external"
                Type:     schema.TypeString,                       // when set: disables existing adapter for interface, creates new adapter for interface with same name as vswitch
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "net_adapter_name" },
            },
        },
    }
}

func resourceHypervVSwitchCreate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    id                            := fmt.Sprintf("//%s/vswitches/%s", host, d.Get("name").(string))
    name                          := d.Get("name").(string)
    switchType                    := strings.ToLower(d.Get("switch_type").(string))
    notes                         := d.Get("notes").(string)
    allowManagementOS             := d.Get("allow_management_os").(bool)
    netAdapterName                := d.Get("net_adapter_name").(string)
    netAdapterInterfaceDesciption := d.Get("net_adapter_interface_description").(string)

    log.Printf(`[INFO][terraform-provider-hyperv] creating hyperv_vswitch %q
                    [INFO][terraform-provider-hyperv]     name:                              %#v
                    [INFO][terraform-provider-hyperv]     switch_type:                       %#v
                    [INFO][terraform-provider-hyperv]     notes:                             %#v
                    [INFO][terraform-provider-hyperv]     allow_management_os:               %#v
                    [INFO][terraform-provider-hyperv]     net_adapter_name:                  %#v
                    [INFO][terraform-provider-hyperv]     net_adapter_interface_description: %#v
`   , id, name, switchType, notes, allowManagementOS, netAdapterName, netAdapterInterfaceDesciption)

    // create vswitch
    vsProperties := new(api.VSwitch)
    vsProperties.Name                           = name
    vsProperties.SwitchType                     = switchType
    vsProperties.Notes                          = notes
    if switchType == "external" {
        vsProperties.AllowManagementOS              = allowManagementOS
        vsProperties.NetAdapterName                 = netAdapterName
        vsProperties.NetAdapterInterfaceDescription = netAdapterInterfaceDesciption
    }

    err := c.CreateVSwitch(vsProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv] cannot create hyperv_vswitch %q\n", id)
        return err
    }

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-hyperv] created hyperv_vswitch %q\n", id)
    return nil
}

func resourceHypervVSwitchRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    id   := d.Id()
    name := d.Get("name").(string)

    log.Printf("[INFO][terraform-provider-hyperv] reading hyperv_vswitch %q\n", id)

    // read vswitch
    vs := new(api.VSwitch)
    vs.Name = name

    vswitch, err := c.ReadVSwitch(vs)
    if err != nil {
        log.Printf("[INFO][terraform-provider-hyperv] cannot read hyperv_vswitch %q\n", id)

        d.SetId("")
        log.Printf("[INFO][terraform-provider-hyperv] deleted hyperv_vswitch %q\n", id)
        return nil   // don't return an error to allow terraform refresh to update state
    }

    // set properties
    d.Set("name", vswitch.Name)
    d.Set("switch_type", strings.ToLower(vswitch.SwitchType))
    d.Set("notes", vswitch.Notes)
    d.Set("allow_management_os", vswitch.AllowManagementOS)
    d.Set("net_adapter_name", vswitch.NetAdapterName)
    d.Set("net_adapter_interface_description", vswitch.NetAdapterInterfaceDescription)

    log.Printf("[INFO][terraform-provider-hyperv] read hyperv_vswitch %q\n", id)
    return nil
}

func resourceHypervVSwitchUpdate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    id                            := d.Id()
    name                          := d.Get("name").(string)
    switchType                    := strings.ToLower(d.Get("switch_type").(string))
    notes                         := d.Get("notes").(string)
    allowManagementOS             := d.Get("allow_management_os").(bool)
    netAdapterName                := d.Get("net_adapter_name").(string)
    netAdapterInterfaceDesciption := d.Get("net_adapter_interface_description").(string)

    log.Printf(`[INFO][terraform-provider-hyperv] updating hyperv_vswitch %q
                    [INFO][terraform-provider-hyperv]     name:                              %#v
                    [INFO][terraform-provider-hyperv]     switch_type:                       %#v
                    [INFO][terraform-provider-hyperv]     notes:                             %#v
                    [INFO][terraform-provider-hyperv]     allow_management_os:               %#v
                    [INFO][terraform-provider-hyperv]     net_adapter_name:                  %#v
                    [INFO][terraform-provider-hyperv]     net_adapter_interface_description: %#v
`   , id, name, switchType, notes, allowManagementOS, netAdapterName, netAdapterInterfaceDesciption)

    // update vswitch
    vs := new(api.VSwitch)
    vs.Name = name

    vsProperties := new(api.VSwitch)
    vsProperties.SwitchType                     = switchType
    vsProperties.Notes                          = notes
    if switchType == "external" {
        vsProperties.AllowManagementOS              = allowManagementOS
        vsProperties.NetAdapterName                 = netAdapterName
        vsProperties.NetAdapterInterfaceDescription = netAdapterInterfaceDesciption
    }

    err := c.UpdateVSwitch(vs, vsProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv] cannot update hyperv_vswitch %q\n", id)
        return err
    }

    log.Printf("[INFO][terraform-provider-hyperv] updated hyperv_vswitch %q\n", id)
    return nil
}

func resourceHypervVSwitchDelete(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    id   := d.Id()
    name := d.Get("name").(string)

    log.Printf("[INFO][terraform-provider-hyperv] deleting hyperv_vswitch %q\n", id)

    // delete vswitch
    vs := new(api.VSwitch)
    vs.Name = name

    err := c.DeleteVSwitch(vs)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv] cannot delete hyperv_vswitch %q\n", id)
        return err
    }

    // set id
    d.SetId("")

    log.Printf("[INFO][terraform-provider-hyperv] deleted hyperv_vswitch %q\n", id)
    return nil
}

func resourceHypervVSwitchImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
    c := m.(*api.HypervClient)

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    importID := d.Id()   // importID is the name of the vswitch
    id       := fmt.Sprintf("//%s/vswitches/%s", host, importID)

    log.Printf("[INFO][terraform-provider-hosts] importing hyperv_vswitch %q\n", id)

    // set id
    d.SetId(id)

    return []*schema.ResourceData{ d }, nil
}

//------------------------------------------------------------------------------
