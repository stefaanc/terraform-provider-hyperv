## PowerShell API Mapping

Mapping terraform resource properties on the PowerShell API.

> :warning:  
> The PowerShell API information in following tables are incomplete PowerShell commands.  The information focusses on the methods and options that are relevant to the related terraform resource property.  In most cases, other options need to be added to make this valid PowerShell code.  

### hyperv_interface

Property                     | PowerShell API
:----------------------------|:-------------- 
`index`                      | `$( Get-NetAdapter ).InterfaceIndex`
`name`                       | `$( Get-NetAdapter ).InterfaceName`
`alias`                      | `$( Get-NetAdapter ).InterfaceAlias`
`description`                | `$( Get-NetAdapter ).InterfaceDescription`
`mac_address`                | `$( Get-NetAdapter ).MacAddress` <br/><br/> alternatively when interface for management-OS <br/><br/> `$( Get-VMNetworkAdapter ).MacAddress`
`network_adapter_name`       | `$( Get-NetAdapter ).Name` <br/><br/> alternatively when interface for management-OS <br/><br/> `$VMNetAdapterName = $( Get-VMNetworkAdapter ).Name` <br/> `"vEthernet (${VMNetAdapterName})"` 
`vnetwork_adapter_name`      | only when interface for management-OS <br/><br/> `$NetAdapterName = $( Get-NetAdapter ).Name` <br/> `$NetAdapterName.Substring(11, $NetAdapterName.Length - 12)` <br/><br/> alternatively <br/><br/> `$( Get-VMNetworkAdapter ).Name` 
`network_name`               | `$( Get-NetConnectionProfile -InterfaceIndex ).Name`
`computer_name`              | `$( Get-NetAdapter ).SystemName`

<br/>

### hyperv_management_os

Property                     | PowerShell API
:----------------------------|:-------------- 
`name`                       | `$env:ComputerName`
`dns`                        | &emsp;
 &emsp; `suffix_search_list` | `Set-DnsClientGlobalSetting -SuffixSearchList`
 &emsp; `enable_devolution`  | `Set-DnsClientGlobalSetting -UseDevolution`
 &emsp; `devolution_level`   | `Set-DnsClientGlobalSetting -DevolutionLevel`

<br/>

### hyperv_network

Property             | PowerShell API
:--------------------|:--------------
`name`               | `$( Get-NetConnectionProfile ).Name`
`connection_profile` | `Set-NetConnectionProfile -NetworkCategory`

<br/>

### hyperv_network_adapter

Property                              | PowerShell API
:-------------------------------------|:--------------
`name`                                | `$( Get-NetAdapter ).Name`
`mac_address`                         | `Set-NetAdapter -MacAddress`
`interface`                           |
 &emsp; `ipv4_interface_disabled`     | `Enable/Disable-NetAdapterBinding -ComponentID ms_tcpip`
 &emsp; `ipv4_interface_metric`       | `Set-NetIPInterface -AddressFamily 'IPv4' -InterfaceMetric`
 &emsp; `ipv6_interface_disabled`     | `Enable/Disable-NetAdapterBinding -ComponentID ms_tcpip6`
 &emsp; `ipv6_interface_metric`       | `Set-NetIPInterface -AddressFamily 'IPv6' -InterfaceMetric`
 &emsp; `register_connection_address` | `Set-DnsClient -RegisterThisConnectionsAddress`
 &emsp; `register_connection_suffix`  | `Set-DnsClient -ConnectionSpecificSuffix -UseSuffixWhenRegistering`
`ip_address`                          | &emsp;
 &emsp; `address`                     | `New-NetIPAddress -IPAddress`
 &emsp; `prefix_length`               | `New-NetIPAddress -PrefixLength`
 &emsp; `skip_as_source`              | `New-NetIPAddress -SkipAsSource`
`gateway`                             | &emsp;
 &emsp; `address`                     | `New-NetRoute -DestinationPrefix '0.0.0.0/0' -NextHop` <br/> or <br/> `New-NetRoute -DestinationPrefix '::/0' -NextHop`
 &emsp; `route_metric`                | `New-NetRoute -DestinationPrefix '0.0.0.0/0' -RouteMetric` <br/> or <br/> `New-NetRoute -DestinationPrefix '::/0' -RouteMetric`
`dns`                                 | `Set-DnsClientServerAddress -ResetServerAddresses` <br/> or <br/> `Set-DnsClientServerAddress -ServerAddresses`
`admin_status`                        | `$( Get-NetAdapter ).AdminStatus`
`operational_status`                  | `$( Get-NetAdapter ).ifOperStatus`
`connection_status`                   | `$( Get-NetAdapter ).MediaConnectionState`
`connection_speed`                    | `$( Get-NetAdapter ).LinkSpeed`
`is_physical`                         | `$( Get-NetAdapter ).ConnectorPresent`

<br/>

### hyperv_vmachine

Property | PowerShell API
:--------|:-------------- 
`name`   |

<br/>

### hyperv_vnetwork_adapter

Property                     | PowerShell API
:----------------------------|:-------------- 
`name`                       | `Add-VMNetworkAdapter -Name`
`vswitch_name`               | `Add-VMNetworkAdapter -SwitchName`
`vmachine_name`              | `Add-VMNetworkAdapter -VMName`
`management_os`              | `Add-VMNetworkAdapter -ManagementOS`
`mac_address`                | `Add-VMNetworkAdapter -DynamicMacAddress` <br/> or <br/> `Add-VMNetworkAdapter -StaticMacAddress`
`allow_mac_address_spoofing` | `Set-VMNetworkAdapter -MacAddressSpoofing` 

<br/>

### hyperv_vswitch

Property                       | PowerShell API
:------------------------------|:-------------- 
`name`                         | `New-VMSwitch -Name`
`uplink`                       | &emsp;
 &emsp; `network_adapter_name` | `New-VMSwitch -SwitchType 'Private'` <br/> or <br/> `New_VMSwitch -NetAdapterName`
 &emsp; `network_adapter_team` | `New-VMSwitch -SwitchType 'Private'` <br/> or <br/> `New-VMSwitch -EnableEmbeddedTeaming` <br/> `Add-VMSwitchTeamMember -NetAdapterName`
`notes`                        | `New-VMSwitch -Notes`

<br/>