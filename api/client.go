//
// Copyright (c) 2019 Sean Reynolds, Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-hyperv
//
package api

import (
)

//------------------------------------------------------------------------------

type HypervClient struct {
    Type       string   // "local" or "ssh"

    // local

    // ssh
    Host       string
    Port       uint16
    User       string
    Password   string
    Insecure   bool
}

//------------------------------------------------------------------------------
