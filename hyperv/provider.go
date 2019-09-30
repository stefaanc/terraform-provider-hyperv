//
// Copyright (c) 2019 Sean Reynolds, Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-hyperv
//
package hyperv

import (
    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
)

func Provider() terraform.ResourceProvider {
    return &schema.Provider{
        },

        ResourcesMap: map[string]*schema.Resource{
        },

        ConfigureFunc: providerConfigure,
    }
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

    config := Config{
    }

    return config.Client()
}
