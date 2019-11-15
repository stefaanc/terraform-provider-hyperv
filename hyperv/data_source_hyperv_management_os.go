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

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

    "github.com/stefaanc/terraform-provider-hyperv/api"
    "github.com/stefaanc/terraform-provider-hyperv/hyperv/tfutil"
)

//------------------------------------------------------------------------------

func dataSourceHypervManagementOS() *schema.Resource {
    return &schema.Resource{
        Read:   dataSourceHypervManagementOSRead,

        Schema: map[string]*schema.Schema{
            "name": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            "dns": &schema.Schema{
                Type:     schema.TypeList,
                Computed: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "suffix_search_list": &schema.Schema{
                            Type:     schema.TypeList,
                            Elem:     &schema.Schema{ Type: schema.TypeString },
                            Computed: true,
                        },
                        "enable_devolution": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },
                        "devolution_level": &schema.Schema{
                            Type:     schema.TypeInt,   // uint32
                            Computed: true,
                        },
                    },
                },
            },
        },
    }
}

//------------------------------------------------------------------------------

func dataSourceHypervManagementOSRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    id := fmt.Sprintf("//%s/management_os", host)

    log.Printf("[INFO][terraform-provider-hyperv] reading hyperv_management_os\n")

    // read management-OS
    managementOS, err := c.ReadManagementOS()
    if err != nil {
        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-hyperv] cannot read hyperv_management_os\n")
        return err
    }

    // set properties
    d.Set("name", managementOS.Name)

    dns := make(map[string]interface{})
    dns["suffix_search_list"] = managementOS.DNS_SuffixSearchList
    dns["enable_devolution"]  = managementOS.DNS_EnableDevolution
    dns["devolution_level"]   = managementOS.DNS_DevolutionLevel
    tfutil.SetResourceDataMap(d, "dns", dns)

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-hyperv] read hyperv_management_os\n")
    return nil
}

//------------------------------------------------------------------------------
