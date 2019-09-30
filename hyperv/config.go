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

type Config struct {
}

func (c *Config) Client() (interface {}, error) {
    log.Printf(`[INFO][terraform-provider-hyperv] configuring hyperv-provider
`)

    hyperv := new(struct{})

    log.Printf("[INFO][terraform-provider-hyperv] configured hyperv-provider\n")
    return hyperv, nil
}
