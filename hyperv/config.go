//
// Copyright (c) 2019 Sean Reynolds, Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-hyperv
//
package hyperv

import (
    "log"

    "github.com/stefaanc/terraform-provider-hyperv/api"
)

//------------------------------------------------------------------------------

type Config struct {
    Type     string

    // ssh
    Host     string
    Port     uint16
    User     string
    Password string
    Insecure bool
}

//------------------------------------------------------------------------------

func (c *Config) Client() (interface {}, error) {
    switch c.Type {
    case "local":
        log.Printf(`[INFO][terraform-provider-hyperv] configuring hyperv-provider
                    [INFO][terraform-provider-hyperv]     type: %q
`       , c.Type)
    case "ssh":
        log.Printf(`[INFO][terraform-provider-hyperv] configuring hyperv-provider
                    [INFO][terraform-provider-hyperv]     type: %q
                    [INFO][terraform-provider-hyperv]     host: %q
                    [INFO][terraform-provider-hyperv]     port: %d
                    [INFO][terraform-provider-hyperv]     user: %q
                    [INFO][terraform-provider-hyperv]     password: ********
                    [INFO][terraform-provider-hyperv]     insecure: %t
`       , c.Type, c.Host, c.Port, c.User, c.Insecure)
    }

    hypervClient := new(api.HypervClient)
    switch c.Type {
    case "local":
        hypervClient.Type     = c.Type
    case "ssh":
        hypervClient.Type     = c.Type
        hypervClient.Host     = c.Host
        hypervClient.Port     = c.Port
        hypervClient.User     = c.User
        hypervClient.Password = c.Password
        hypervClient.Insecure = c.Insecure
    }

    log.Printf("[INFO][terraform-provider-hyperv] configured hyperv-provider\n")
    return hypervClient, nil
}

//------------------------------------------------------------------------------
