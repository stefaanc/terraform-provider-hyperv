//
// Copyright (c) 2019 Sean Reynolds, Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-hyperv
//
package tfutil

import (
    "strings"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

//------------------------------------------------------------------------------

func StateToLower() schema.SchemaStateFunc {
    return func(val interface{}) string {
        return strings.ToLower(val.(string))
    }
}

//------------------------------------------------------------------------------

func DiffSuppressCase() schema.SchemaDiffSuppressFunc {
    return func(k, old, new string, d *schema.ResourceData) bool {
        if strings.ToLower(old) == strings.ToLower(new) {
            return true
        }
        return false
    }
}

//------------------------------------------------------------------------------
