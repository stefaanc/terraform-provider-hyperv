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

    "github.com/stefaanc/terraform-provider-hyperv/api"
    "github.com/stefaanc/terraform-provider-hyperv/hyperv/tfutil"
)

//------------------------------------------------------------------------------

func dataSourceHypervNetworkAdapter() *schema.Resource {
    return &schema.Resource{
        Read:   dataSourceHypervNetworkAdapterRead,

        Schema: map[string]*schema.Schema{
            "name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
            },

            "mac_address": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },

            "interface": &schema.Schema{
                Type:     schema.TypeList,
                Computed: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "ipv4_interface_disabled": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },
                        "ipv4_interface_metric": &schema.Schema{
                            Type:     schema.TypeInt,
                            Computed: true,
                        },
                        "ipv6_interface_disabled": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },
                        "ipv6_interface_metric": &schema.Schema{
                            Type:     schema.TypeInt,
                            Computed: true,
                        },
                        "register_connection_address": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },
                        "register_connection_suffix": &schema.Schema{
                            Type:     schema.TypeString,
                            Computed: true,
                        },
                    },
                },
            },

            "ip_address": &schema.Schema{
                Type:     schema.TypeSet,
                Computed: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "address": &schema.Schema{
                            Type:     schema.TypeString,
                            Computed: true,
                        },
                        "prefix_length": &schema.Schema{
                            Type:     schema.TypeInt,
                            Computed: true,
                        },
                        "skip_as_source": &schema.Schema{
                            Type:     schema.TypeBool,
                            Computed: true,
                        },
                    },
                },
            },

            "gateway": &schema.Schema{
                Type:     schema.TypeSet,
                Computed: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "address": &schema.Schema{
                            Type:     schema.TypeString,
                            Computed: true,
                        },
                        "route_metric": &schema.Schema{
                            Type:     schema.TypeInt,
                            Computed: true,
                        },
                    },
                },
            },

            "dns": &schema.Schema{
                Type:     schema.TypeSet,
                Computed: true,
                Elem: &schema.Schema{ Type: schema.TypeString },
            },

            "admin_status": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "operational_status": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "connection_status": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "connection_speed": &schema.Schema{
                Type:     schema.TypeString,
                Computed: true,
            },
            "is_physical": &schema.Schema{
                Type:     schema.TypeBool,
                Computed: true,
            },

            // lifecycle customizations that are not supported by the 'lifecycle' meta-argument for data sources
            "x_lifecycle": &tfutil.DataSourceXLifecycleSchema,
        },
    }
}

//------------------------------------------------------------------------------

func dataSourceHypervNetworkAdapterRead(d *schema.ResourceData, m interface{}) error {
    c := m.(*api.HypervClient)

    name        := d.Get("name").(string)
    x_lifecycle := tfutil.GetResourceDataMap(d, "x_lifecycle")

    host := "localhost"
    if c.Type != "local" {
        host = c.Host
    }

    id := fmt.Sprintf("//%s/network_adapters/%s", host, name)

    log.Printf("[INFO][terraform-provider-hyperv] reading hyperv_network_adapter %q\n", id)

    // read network_adapter
    naQuery := new(api.NetworkAdapter)
    naQuery.Name = name

    networkAdapter, err := c.ReadNetworkAdapter(naQuery)
    if err != nil {
        // lifecycle customizations: ignore_error_if_not_exists
        if x_lifecycle != nil {
            ignore_error_if_not_exists := x_lifecycle["ignore_error_if_not_exists"].(bool)
            if ignore_error_if_not_exists && strings.Contains(err.Error(), "cannot find network_adapter") {
                log.Printf("[INFO][terraform-provider-hyperv] cannot read hyperv_network_adapter %q\n", id)

                // set zeroed properties
                d.Set("name", "")
                d.Set("mac_address", "")

                ip_interface := make(map[string]interface{})
                ip_interface["ipv4_interface_disabled"] = false
                ip_interface["ipv4_interface_metric"] = 0
                ip_interface["ipv6_interface_disabled"] = false
                ip_interface["ipv6_interface_metric"] = 0
                ip_interface["register_connection_address"] = false
                ip_interface["register_connection_suffix"] = ""
                tfutil.SetResourceDataMap(d, "interface", ip_interface)

                ipAddressResource := dataSourceHypervNetworkAdapter().Schema["ip_address"].Elem.(*schema.Resource)
                d.Set("ip_address", schema.NewSet(schema.HashResource(ipAddressResource), nil))
                gatewayResource := dataSourceHypervNetworkAdapter().Schema["gateway"].Elem.(*schema.Resource)
                d.Set("gateway", schema.NewSet(schema.HashResource(gatewayResource), nil))
                d.Set("dns", schema.NewSet(schema.HashString, nil))

                d.Set("admin_status", "")
                d.Set("operational_status", "")
                d.Set("connection_status", "")
                d.Set("connection_speed", "")
                d.Set("is_physical", false)

                // set computed lifecycle properties
                x_lifecycle["exists"] = false
                tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)

                // set id
                d.SetId(id)

                log.Printf("[INFO][terraform-provider-hyperv] ignored error and added zeroed hyperv_network_adapter %q to terraform state\n", id)
                return nil
            }
        }
        // no lifecycle customizations
        log.Printf("[ERROR][terraform-provider-hyperv] cannot read hyperv_network_adapter %q\n", id)
        return err
    }

    // set properties
    d.Set("name", networkAdapter.Name)
    d.Set("mac_address", networkAdapter.MACAddress)

    ip_interface := make(map[string]interface{})
    ip_interface["ipv4_interface_disabled"] = networkAdapter.IPv4InterfaceDisabled
    ip_interface["ipv4_interface_metric"] = networkAdapter.IPv4InterfaceMetric
    ip_interface["ipv6_interface_disabled"] = networkAdapter.IPv6InterfaceDisabled
    ip_interface["ipv6_interface_metric"] = networkAdapter.IPv6InterfaceMetric
    ip_interface["register_connection_address"] = networkAdapter.RegisterConnectionAddress
    ip_interface["register_connection_suffix"] = networkAdapter.RegisterConnectionSuffix
    tfutil.SetResourceDataMap(d, "interface", ip_interface)

    ipAddresses := make([]interface{}, len(networkAdapter.IPAddresses))
    for i, a := range networkAdapter.IPAddresses {
        address := make(map[string]interface{})
        address["address"] = a.Address
        address["prefix_length"] = a.PrefixLength
        address["skip_as_source"] = a.SkipAsSource
        ipAddresses[i] = address
    }
    ipAddressResource := dataSourceHypervNetworkAdapter().Schema["ip_address"].Elem.(*schema.Resource)
    d.Set("ip_address", schema.NewSet(schema.HashResource(ipAddressResource), ipAddresses))

    gateways := make([]interface{}, len(networkAdapter.Gateways))
    for i, g := range networkAdapter.Gateways {
        gateway := make(map[string]interface{})
        gateway["address"] = g.Address
        gateway["route_metric"] = g.RouteMetric
        gateways[i] = gateway
    }
    gatewayResource := dataSourceHypervNetworkAdapter().Schema["gateway"].Elem.(*schema.Resource)
    d.Set("gateway", schema.NewSet(schema.HashResource(gatewayResource), gateways))

    dnservers := make([]interface{}, len(networkAdapter.DNServers))
    for i, s := range networkAdapter.DNServers {
        dnservers[i] = s
    }
    d.Set("dns", schema.NewSet(schema.HashString, dnservers))

    d.Set("admin_status", networkAdapter.AdminStatus)
    d.Set("operational_status", networkAdapter.OperationalStatus)
    d.Set("connection_status", networkAdapter.ConnectionStatus)
    d.Set("connection_speed", networkAdapter.ConnectionSpeed)
    d.Set("is_physical", networkAdapter.IsPhysical)

    // set computed lifecycle properties
    if x_lifecycle != nil {
        x_lifecycle["exists"] = true
        tfutil.SetResourceDataMap(d, "x_lifecycle", x_lifecycle)
    }

    // set id
    d.SetId(id)

    log.Printf("[INFO][terraform-provider-hyperv] read hyperv_network_adapter %q\n", id)
    return nil
}

//------------------------------------------------------------------------------
