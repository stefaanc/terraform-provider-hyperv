###
### hyperv
###

provider "hyperv" {
    version = "~> 0.0.0"
    alias = "local"

    type = "local"
}

data "hyperv_interface" "test" {
    provider = hyperv.local

    vnetwork_adapter_name = "External Switch"
}

data "hyperv_management_os" "local" {
    provider = hyperv.local
}

data "hyperv_network" "network" {
    provider = hyperv.local

    name = "network"
}

data "hyperv_network_adapter" "external" {
    provider = hyperv.local

    name = "vEthernet (External Switch)"
}

# data "hyperv_vswitch" "vs_default" {
#     provider = hyperv.local

#     name = "Default Switch"

#     x_lifecycle {
#         ignore_error_if_not_exists = true
#     }
# }

# resource "hyperv_vswitch" "vs_private" {
#     provider = hyperv.local

#     name                = "Private Switch"
#     switch_type         = "private"
#     notes               = "private notes"
# }

# resource "hyperv_vswitch" "vs_internal" {
#     provider = hyperv.local

#     name                = "Internal Switch"
#     switch_type         = "internal"
#     notes               = "internal notes"
# }

# resource "hyperv_vswitch" "vs_external" {
#     provider = hyperv.local

#     name                = "External Switch"
#     switch_type         = "external"
#     notes               = "external notes"

#     allow_management_os = true
#     #net_adapter_name    = "Ethernet"
#     net_adapter_interface_description = "Intel(R) 82579LM Gigabit Network Connection"

#     x_lifecycle {
#         import_if_exists = true
#     }
# }
