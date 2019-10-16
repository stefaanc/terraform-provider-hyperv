###
### hyperv
###

provider "hyperv" {
    version = "~> 0.0.0"
    alias = "local"

    type = "local"
}

data "hyperv_vswitch" "defaultvs" {
    provider = hyperv.local

    name = "Default Switch"
}

resource "hyperv_vswitch" "privatevs" {
    provider = hyperv.local

    name                = "Private Switch"
    switch_type         = "private"
    notes               = "private notes"
}

resource "hyperv_vswitch" "internalvs" {
    provider = hyperv.local

    name                = "Internal Switch"
    switch_type         = "internal"
    notes               = "internal notes"
}

# resource "hyperv_vswitch" "externalvs" {
#     provider = hyperv.local
#
#     name                = "External"
#     switch_type         = "external"
#     notes               = "external notes"
#     allow_management_os = true
#     net_adapter_name    = "Ethernet"
# #    net_adapter_interface_description ="Intel(R) 82579LM Gigabit Network Connection"
# }
