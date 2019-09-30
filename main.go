//
// Copyright (c) 2019 Sean Reynolds, Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-hyperv
//
package main

import (
    "github.com/hashicorp/terraform-plugin-sdk/plugin"
    
    "github.com/stefaanc/terraform-provider-hyperv/hyperv"
)

func main() {
    plugin.Serve(&plugin.ServeOpts{
        ProviderFunc: hosts.Provider,
    })
}