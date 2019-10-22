//
// Copyright (c) 2019 Sean Reynolds, Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-hyperv
//
package hyperv

import (
    "strings"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/terraform"
    "github.com/hashicorp/terraform-plugin-sdk/helper/validation"

    "github.com/stefaanc/terraform-provider-hyperv/hyperv/tfutil"
)

//------------------------------------------------------------------------------

func Provider() terraform.ResourceProvider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema {
            "type": &schema.Schema{
                Description: "The type of connection to the hyperv-server: \"local\" or \"ssh\"",
                Type:     schema.TypeString,
                Optional: true,
                Default: "local",

                ValidateFunc:     validation.StringInSlice([]string{ "local", "ssh" }, true),
                StateFunc:        tfutil.StateToLower(),
                DiffSuppressFunc: tfutil.DiffSuppressCase(),
            },

            // ssh
            "host": &schema.Schema{                                // config ignored when type is not "ssh"
                Description: "The hyperv-server",
                Type:     schema.TypeString,
                Optional: true,
                Default: "localhost",

                DiffSuppressFunc: tfutil.DiffSuppressCase(),
            },
            "port": &schema.Schema{                                // config ignored when type is not "ssh"
                Description: "The hyperv-server's port for ssh",
                Type:     schema.TypeInt,
                Optional: true,
                Default:  22,

                ValidateFunc: validation.IntBetween(0, 65535),
            },
            "user": &schema.Schema{                                // config ignored when type is not "ssh"
                Description: "The user name for communication with the hyperv-server",
                Type:     schema.TypeString,
                Optional: true,
                Default: "",
            },
            "password": &schema.Schema{                            // config ignored when type is not "ssh"
                Description: "The user password for communication with the hyperv-server",
                Type:      schema.TypeString,
                Optional:  true,
                Default:   "",
                Sensitive: true,
            },
            "insecure": &schema.Schema{                            // config ignored when type is not "ssh"
                Description: "Allow insecure communication - disables checking of the server certificate",
                Type:     schema.TypeBool,
                Optional: true,
                Default: false,
            },
        },

        DataSourcesMap: map[string]*schema.Resource {
            "hyperv_vswitch": dataSourceHypervVSwitch(),
        },

        ResourcesMap: map[string]*schema.Resource{
            "hyperv_vswitch": resourceHypervVSwitch(),
        },

        ConfigureFunc: providerConfigure,
    }
}

//------------------------------------------------------------------------------

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
    config := Config{
        Type:     strings.ToLower(d.Get("type").(string)),

        // ssh
        Host:     d.Get("host").(string),
        Port:     uint16(d.Get("port").(int)),
        User:     d.Get("user").(string),
        Password: d.Get("password").(string),
        Insecure: d.Get("insecure").(bool),
    }

    return config.Client()
}

//------------------------------------------------------------------------------
