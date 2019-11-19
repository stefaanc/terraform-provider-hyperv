###
### hyperv
###

provider "hyperv" {
    type = "local"
}

data "hyperv_interface" "test" {
    vnetwork_adapter_name = "External Switch"
}

data "hyperv_management_os" "local" {
}

data "hyperv_network" "network" {
    name = "network"
}

data "hyperv_network_adapter" "external" {
    name = "vEthernet (External Switch)"
}

data "hyperv_vnetwork_adapter" "external" {
    name = "External Switch"
}

data "hyperv_vnetwork_adapter" "test" {
    name          = "Network Adapter"
    vmachine_name = "Video"
}

data "hyperv_vswitch" "default" {
    name = "Default Switch"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}

data "hyperv_vswitch" "external" {
     name = "External Switch"
}

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
