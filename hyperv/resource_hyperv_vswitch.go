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

            // config when switch_type is "external"
            "allow_management_os": &schema.Schema{                 // defaults to true when switch_type is "internal", to false when switch_type is "private"
                Type:     schema.TypeBool,
                Optional: true,
                Computed: true,
            },
            "net_adapter_name": &schema.Schema{                    // defaults to "" when switch_type is "private" or "internal"
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                DiffSuppressFunc: tfutil.DiffSuppressCase(),
            },
            "net_adapter_interface_description": &schema.Schema{   // defaults to "" when switch_type is "private" or "internal"
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,

                ConflictsWith: []string{ "net_adapter_name" },
            },

            // lifecycle customizations that are not supported by the 'lifecycle' meta-argument for resources
            "x_lifecycle": &tfutil.ResourceXLifecycleSchema,
                // remark that as a general rule, "import_if_exists" will fail if any of the properties in the config are not the same as the properties of existing resource
                // exception to this rule: when only the "notes" property is different, the existing switch will be imported and updated
        },

        CustomizeDiff: validateConflictsWithSwitchType,
    }
}

func validateConflictsWithSwitchType(diff *schema.ResourceDiff, m interface{}) error {
    switch_type := strings.ToLower(diff.Get("switch_type").(string))

    // "allow_management_os"
    if switch_type == "private" {
        if v, ok := diff.GetOkExists("allow_management_os"); ok && v.(bool) {
            return fmt.Errorf("\"allow_management_os\": conflicts with 'switch_type = %q'", switch_type)
        }
    }
    if switch_type == "internal" {
        if v, ok := diff.GetOkExists("allow_management_os"); ok && !v.(bool) {
            return fmt.Errorf("\"allow_management_os\": conflicts with 'switch_type = %q'", switch_type)
        }
    }

    // "net_adapter_name" and "net_adapter_interface_description"
    if switch_type == "private" || switch_type == "internal" {
        if diff.Get("net_adapter_name").(string) != "" {
            return fmt.Errorf("\"net_adapter_name\": conflicts with 'switch_type = %q'", switch_type)
        }
        if diff.Get("net_adapter_interface_description").(string) != "" {
            return fmt.Errorf("\"net_adapter_interface_description\": conflicts with 'switch_type = %q'", switch_type)
        }
    }
    return nil
}

func resourceHypervVSwitchCreate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    id                             := fmt.Sprintf("//%s/vswitches/%s", host, d.Get("name").(string))
    name                           := d.Get("name").(string)
    switchType                     := strings.ToLower(d.Get("switch_type").(string))
    notes                          := d.Get("notes").(string)
    allowManagementOS              := d.Get("allow_management_os").(bool)
    netAdapterName                 := d.Get("net_adapter_name").(string)
    netAdapterInterfaceDescription := d.Get("net_adapter_interface_description").(string)
    x_lifecycle                    := tfutil.GetResourceDataMap(d, "x_lifecycle")

    allowManagementOS_msg              := d.Get("allow_management_os")
    netAdapterName_msg                 := d.Get("net_adapter_name")
    netAdapterInterfaceDescription_msg := d.Get("net_adapter_interface_description")
    if _, ok := d.GetOkExists("allowManagementOS"); !ok { allowManagementOS_msg              = "(computed)" }
    if netAdapterName == ""                             { netAdapterName_msg                 = "(computed)" }
    if netAdapterInterfaceDescription == ""             { netAdapterInterfaceDescription_msg = "(computed)" }
    log.Printf(`[INFO][terraform-provider-hyperv] creating hyperv_vswitch %q
                    [INFO][terraform-provider-hyperv]     name:                              %#v
                    [INFO][terraform-provider-hyperv]     switch_type:                       %#v
                    [INFO][terraform-provider-hyperv]     notes:                             %#v
                    [INFO][terraform-provider-hyperv]     allow_management_os:               %#v
                    [INFO][terraform-provider-hyperv]     net_adapter_name:                  %#v
                    [INFO][terraform-provider-hyperv]     net_adapter_interface_description: %#v
`   , id, name, switchType, notes, allowManagementOS_msg, netAdapterName_msg, netAdapterInterfaceDescription_msg)

    // create vswitch
    vsProperties := new(api.VSwitch)
    vsProperties.Name                           = name
    vsProperties.SwitchType                     = switchType
    vsProperties.Notes                          = notes
    if switchType == "external" {
        vsProperties.AllowManagementOS              = allowManagementOS
        vsProperties.NetAdapterName                 = netAdapterName
        vsProperties.NetAdapterInterfaceDescription = netAdapterInterfaceDescription
    }

    err := c.CreateVSwitch(vsProperties)
    if err != nil {
        // lifecycle customizations: import_if_exists
        if x_lifecycle != nil {
            import_if_exists := x_lifecycle["import_if_exists"].(bool)
            if import_if_exists && strings.Contains(err.Error(), "already exists") {
                log.Printf("[INFO][terraform-provider-hyperv] cannot create hyperv_vswitch %q\n", id)
                log.Printf("[INFO][terraform-provider-hyperv] importing hyperv_vswitch %q into terraform state\n", id)

                // read vswitch
                vs := new(api.VSwitch)
                vs.Name = name

                vswitch, err := c.ReadVSwitch(vs)
                if err != nil {
                    log.Printf("[ERROR][terraform-provider-hyperv] cannot read existing hyperv_vswitch %q\n", id)
                    log.Printf("[ERROR][terraform-provider-hyperv] cannot import hyperv_vswitch %q into terraform state\n", id)
                    return err
                }

                if vswitch.SwitchType != switchType ||
                   ( vswitch.SwitchType == "external" &&
                     ( vswitch.AllowManagementOS != allowManagementOS ||
                       ( netAdapterName != "" && vswitch.NetAdapterName != netAdapterName ) ||
                       ( netAdapterName == "" && vswitch.NetAdapterInterfaceDescription != netAdapterInterfaceDescription ) ) ) {
                    err = fmt.Errorf("[terraform-provider-hyperv/hyperv/resourceHypervVSwitchCreate()] cannot import hyperv_vswitch %q into terraform state when terraform config doesn't match the properties in infrastructure", name)

                    log.Printf(`[ERROR][terraform-provider-hyperv] terraform config for hyperv_vswitch %q doesn't match the existing properties
                        [ERROR][terraform-provider-hyperv]     name:                              %#v
                        [ERROR][terraform-provider-hyperv]     switch_type:                       %#v
                        [ERROR][terraform-provider-hyperv]     notes:                             %#v
                        [ERROR][terraform-provider-hyperv]     allow_management_os:               %#v
                        [ERROR][terraform-provider-hyperv]     net_adapter_name:                  %#v
                        [ERROR][terraform-provider-hyperv]     net_adapter_interface_description: %#v
`                   , id, vswitch.Name, vswitch.SwitchType, vswitch.Notes, vswitch.AllowManagementOS, vswitch.NetAdapterName, vswitch.NetAdapterInterfaceDescription)
                    log.Printf("[ERROR][terraform-provider-hyperv] cannot import hyperv_vswitch %q into terraform state\n", id)
                    return err
                }

                // update vswitch
                if vswitch.Notes != notes {
                    err := c.UpdateVSwitch(vs, vsProperties)
                    if err != nil {
                        log.Printf("[ERROR][terraform-provider-hyperv] cannot update existing hyperv_vswitch %q\n", id)
                        log.Printf("[ERROR][terraform-provider-hyperv] cannot import hyperv_vswitch %q into terraform state\n", id)
                        return err
                    }
                }

                // set computed lifecycle properties
                x_lifecycle["imported"] = true
                tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)

                // set id
                d.SetId(id)

                log.Printf("[INFO][terraform-provider-hyperv] imported hyperv_vswitch %q into terraform state\n", id)
                return resourceHypervVSwitchRead(d, m)
            }
        }

        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-hyperv] cannot create hyperv_vswitch %q\n", id)
        return err
    }

    // set computed lifecycle properties
    if x_lifecycle != nil {
        x_lifecycle["imported"] = false
        tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)
    }

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-hyperv] created hyperv_vswitch %q\n", id)
    return resourceHypervVSwitchRead(d, m)
}

func resourceHypervVSwitchRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    id          := d.Id()
    name        := d.Get("name").(string)
    x_lifecycle := tfutil.GetResourceDataMap(d, "x_lifecycle")

    log.Printf("[INFO][terraform-provider-hyperv] reading hyperv_vswitch %q\n", id)

    // read vswitch
    vs := new(api.VSwitch)
    vs.Name = name

    vswitch, err := c.ReadVSwitch(vs)
    if err != nil {
        log.Printf("[INFO][terraform-provider-hyperv] cannot read hyperv_vswitch %q\n", id)

        // set id
        d.SetId("")

        log.Printf("[INFO][terraform-provider-hyperv] deleted hyperv_vswitch %q from terraform state\n", id)
        return nil   // don't return an error to allow terraform refresh to update state
    }

    // set properties
    d.Set("name", vswitch.Name)
    d.Set("switch_type", strings.ToLower(vswitch.SwitchType))
    d.Set("notes", vswitch.Notes)
    d.Set("allow_management_os", vswitch.AllowManagementOS)
    d.Set("net_adapter_name", vswitch.NetAdapterName)
    d.Set("net_adapter_interface_description", vswitch.NetAdapterInterfaceDescription)
    tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)   // make sure new terraform state includes 'x_lifecycle' from the old terraform state when doing a terraform refresh

    log.Printf("[INFO][terraform-provider-hyperv] read hyperv_vswitch %q\n", id)
    return nil
}

func resourceHypervVSwitchUpdate(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    id                             := d.Id()
    name                           := d.Get("name").(string)
    switchType                     := strings.ToLower(d.Get("switch_type").(string))
    notes                          := d.Get("notes").(string)
    allowManagementOS              := d.Get("allow_management_os").(bool)
    netAdapterName                 := d.Get("net_adapter_name").(string)
    netAdapterInterfaceDescription := d.Get("net_adapter_interface_description").(string)

    log.Printf(`[INFO][terraform-provider-hyperv] updating hyperv_vswitch %q
                    [INFO][terraform-provider-hyperv]     name:                              %#v
                    [INFO][terraform-provider-hyperv]     switch_type:                       %#v
                    [INFO][terraform-provider-hyperv]     notes:                             %#v
                    [INFO][terraform-provider-hyperv]     allow_management_os:               %#v
                    [INFO][terraform-provider-hyperv]     net_adapter_name:                  %#v
                    [INFO][terraform-provider-hyperv]     net_adapter_interface_description: %#v
`   , id, name, switchType, notes, allowManagementOS, netAdapterName, netAdapterInterfaceDescription)

    // changes in 'x_lifecycle' only, must not trigger an update in infrastructure
    if !d.HasChange("switch_type") &&   // strictly speaking, this is not required since 'ForceNew = true', but we add this in case we change to 'ForceNew = false'
       !d.HasChange("notes") &&
       !d.HasChange("allow_management_os") &&
       !d.HasChange("net_adapter_name") &&
       !d.HasChange("net_adapter_interface_description") {
        log.Printf("[INFO][terraform-provider-hyperv] updated hyperv_vswitch %q in terraform state, no change in infrastructure\n", id)
        return resourceHypervVSwitchRead(d, m)
    }

    // update vswitch
    vs := new(api.VSwitch)
    vs.Name = name

    vsProperties := new(api.VSwitch)
    vsProperties.SwitchType                     = switchType
    vsProperties.Notes                          = notes
    if switchType == "external" {
        vsProperties.AllowManagementOS              = allowManagementOS
        vsProperties.NetAdapterName                 = netAdapterName
        vsProperties.NetAdapterInterfaceDescription = netAdapterInterfaceDescription
    }

    err := c.UpdateVSwitch(vs, vsProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv] cannot update hyperv_vswitch %q\n", id)
        return err
    }

    log.Printf("[INFO][terraform-provider-hyperv] updated hyperv_vswitch %q\n", id)
    return resourceHypervVSwitchRead(d, m)
}

func resourceHypervVSwitchDelete(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    id          := d.Id()
    name        := d.Get("name").(string)
    x_lifecycle := tfutil.GetResourceDataMap(d, "x_lifecycle")

    log.Printf("[INFO][terraform-provider-hyperv] deleting hyperv_vswitch %q\n", id)

    // lifecycle customizations: destroy_if_imported
    if x_lifecycle != nil {
        imported := x_lifecycle["imported"].(bool)
        destroy_if_imported := x_lifecycle["destroy_if_imported"].(bool)
        if imported && !destroy_if_imported {
            log.Printf("[INFO][terraform-provider-hyperv] hyperv_vswitch %q was imported and must not be deleted from infrastructure\n", id)

            // set id
            d.SetId("")

            log.Printf("[INFO][terraform-provider-hyperv] deleted hyperv_vswitch %q from terraform state, no change in infrastructure\n", id)
            return nil
        }
    }

    // no lifecycle customizations
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
