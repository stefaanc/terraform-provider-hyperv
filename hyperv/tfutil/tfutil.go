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

var DataSourceXLifecycleSchema schema.Schema =  schema.Schema{
    Type:     schema.TypeList,
    Optional: true,
    MaxItems: 1,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            // "ignore_error_if_not_exists" ignores the error when the data source doesn't exist, the resource is added to the terraform state with zeroed properties
            // this can be used in conditional resources
            "ignore_error_if_not_exists": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
                Default: false,
            },
            // "exists" true if the resource exists
            "exists": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
        },
    },
}

var ResourceXLifecycleSchema schema.Schema = schema.Schema{
    Type:     schema.TypeList,
    Optional: true,
    MaxItems: 1,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            // "import_if_exists" imports the resource into the terraform state when creating a resource that already exists
            // this will fail if any of the properties in the config are not the same as the properties of existing resource - this is to reduce the risk of accidental imports
            // this can be used to adapt resources that are shared with external systems
            "import_if_exists": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
                Default: false,
            },
            // "imported" set if the resource was imported using the "import_if_exists" lifecycle customization
            "imported": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },
            // "destroy_if_imported" destroys the resource from the infrastructure when using 'terraform destroy' and when it was imported using 'import_if_exists = true'
            // by default, a resource that is imported using 'import_if_exists = true' is not destroyed from the infrastructure when using 'terraform destroy'
            "destroy_if_imported": &schema.Schema{
                Type:     schema.TypeBool,
                Optional: true,
                Default: false,
            },
        },
    },
}

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

func GetResourceDataMap(d *schema.ResourceData, name string) (m map[string]interface{}) {
    list := d.Get(name).([]interface{})
    if len(list) > 0 {
        m = list[0].(map[string]interface{})
    }
    return m
}

func SetResourceDataMap(d *schema.ResourceData, name string, m map[string]interface{}) error {
    if len(m) == 0 {
        return d.Set(name, []map[string]interface{}(nil))
    }
    return d.Set(name, []map[string]interface{}{m})
}

//------------------------------------------------------------------------------
