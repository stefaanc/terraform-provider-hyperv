//
// Copyright (c) 2019 Sean Reynolds, Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-hyperv
//
package api

import (
    "bytes"
    "errors"
    "fmt"
    "encoding/json"
    "log"
    "strings"

    "github.com/stefaanc/golang-exec/runner"
    "github.com/stefaanc/golang-exec/script"
)

//------------------------------------------------------------------------------

type NetworkAdapter struct {
    Name                      string
    MACAddress                string

    // interface
    IPv4InterfaceDisabled     bool
    IPv4InterfaceMetric       uint32
    IPv6InterfaceDisabled     bool
    IPv6InterfaceMetric       uint32
    RegisterConnectionAddress bool
    RegisterConnectionSuffix  string

    // IP addresses
    IPAddresses               []IPAddress

    // gateways
    Gateways                  []Gateway

    // DNServers
    DNServers                 []string

    // status
    AdminStatus               string
    OperationalStatus         string
    ConnectionStatus          string
    ConnectionSpeed           string
    IsPhysical                bool
}

type IPAddress struct {
    Address      string
    PrefixLength uint8
    SkipAsSource bool
}

type Gateway struct {
    Address     string
    RouteMetric uint16
}

//------------------------------------------------------------------------------

func (c *HypervClient) ReadNetworkAdapter(naQuery *NetworkAdapter) (naProperties *NetworkAdapter, err error) {
    if naQuery.Name == "" {
        return nil, fmt.Errorf("[ERROR][terraform-provider-hyperv/api/ReadNetworkAdapter(naQuery)] missing 'naQuery.Name'")
    }

    return readNetworkAdapter(c, naQuery)
}

//------------------------------------------------------------------------------

func readNetworkAdapter(c *HypervClient, naQuery *NetworkAdapter) (naProperties *NetworkAdapter, err error) {
    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, readNetworkAdapterScript, readNetworkAdapterArguments{
        Name: naQuery.Name,
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetworkAdapter()] cannot read network_adapter %#v\n", naQuery.Name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetworkAdapter()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetworkAdapter()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetworkAdapter()] script stderr: %s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-hyperv/api/readNetworkAdapter()] runner: %s", stderr.String())
        }

        return nil, err
    }

    // convert stdout-JSON to naProperties
    naProperties = new(NetworkAdapter)
    err = json.Unmarshal(stdout.Bytes(), naProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetworkAdapter()] cannot convert json to 'naProperties' for network_adapter %#v\n", naQuery.Name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetworkAdapter()] json: %s", stdout.String())
        return nil, err
    }

    log.Printf("[INFO][terraform-provider-hyperv/api/readNetworkAdapter()] read network_adapter %#v\n", naQuery.Name)
    return naProperties, nil
}

type readNetworkAdapterArguments struct{
    Name string
}

var readNetworkAdapterScript = script.New("readNetworkAdapter", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $networkAdapter = Get-NetAdapter -Name '{{.Name}}' -ErrorAction 'Ignore'
    if ( -not $networkAdapter ) {
        throw "cannot find network_adapter '{{.Name}}'"
    }

    # prepare result
    $naProperties = @{
        Name              = $networkAdapter.Name
        MACAddress        = $networkAdapter.MacAddress

        IPAddresses       = @()
        Gateways          = @()
        DNServers         = @()

        AdminStatus       = $networkAdapter.AdminStatus.ToString()
        OperationalStatus = $networkAdapter.ifOperStatus.ToString()
        ConnectionStatus  = $networkAdapter.MediaConnectionState.ToString()
        ConnectionSpeed   = $networkAdapter.LinkSpeed.ToString()
        IsPhysical        = $networkAdapter.ConnectorPresent
    }

    $ipv4Binding = Get-NetAdapterBinding -Name '{{.Name}}' -ComponentID 'ms_tcpip' -ErrorAction 'Ignore'
    if ( $ipv4Binding ) {
        $naProperties.IPv4InterfaceDisabled = -not $ipv4Binding.Enabled
    }

    $ipv4Interface = Get-NetIPInterface -InterfaceIndex $networkAdapter.InterfaceIndex -AddressFamily 'IPv4' -ErrorAction 'Ignore'
    if ( $ipv4Interface ) {
        $naProperties.IPv4InterfaceMetric = $ipv4Interface.IPv4InterfaceMetric
    }

    $ipv6Binding = Get-NetAdapterBinding -Name '{{.Name}}' -ComponentID 'ms_tcpip6' -ErrorAction 'Ignore'
    if ( $ipv6Binding ) {
        $naProperties.IPv6InterfaceDisabled = -not $ipv6Binding.Enabled
    }

    $ipv6Interface = Get-NetIPInterface -InterfaceIndex $networkAdapter.InterfaceIndex -AddressFamily 'IPv6' -ErrorAction 'Ignore'
    if ( $ipv6Interface ) {
        $naProperties.IPv6InterfaceMetric = $ipv6Interface.IPv6InterfaceMetric
    }

    $dnsClient = Get-DNSClient -InterfaceIndex $networkAdapter.InterfaceIndex -ErrorAction 'Ignore'
    if ( $dnsClient ) {
        $naProperties.RegisterConnectionAddress = $dnsClient.RegisterThisConnectionsAddress
        if ( $dnsClient.UseSuffixWhenRegistering ) {
            $naProperties.RegisterConnectionSuffix  = $dnsClient.ConnectionSpecificSuffix
        }
    }

    $ipv4Address = Get-NetIPAddress -InterfaceIndex $networkAdapter.InterfaceIndex -AddressFamily 'IPv4' -AddressState 'Preferred' -ErrorAction 'Ignore'
    if ( $ipv4Address ) {
        $ipv4Address | foreach {
            $naProperties.IPAddresses += @{
                Address      = $_.IPAddress
                PrefixLength = $_.PrefixLength
                SkipAsSource = $_.SkipAsSource
            }
        }
    }

    $ipv6Address = Get-NetIPAddress -InterfaceIndex $networkAdapter.InterfaceIndex -AddressFamily 'IPv6' -AddressState 'Preferred' -ErrorAction 'Ignore'
    if ( $ipv6Address ) {
        $ipv6Address | foreach {
            $naProperties.IPAddresses += @{
                Address      = $_.IPAddress
                PrefixLength = $_.PrefixLength
                SkipAsSource = $_.SkipAsSource
            }
        }
    }

    $ipv4Route = Get-NetRoute -InterfaceIndex $networkAdapter.InterfaceIndex -AddressFamily 'IPv4' -DestinationPrefix '0.0.0.0/0' -ErrorAction 'Ignore'
    if ( $ipv4Route ) {
        $ipv4Route | foreach {
            $naProperties.Gateways += @{
                Address     = $_.NextHop
                RouteMetric = $_.RouteMetric
            }
        }
    }

    $ipv6Route = Get-NetRoute -InterfaceIndex $networkAdapter.InterfaceIndex -AddressFamily 'IPv6' -DestinationPrefix '::/0' -ErrorAction 'Ignore'
    if ( $ipv6Route ) {
        $ipv6Route | foreach {
            $naProperties.Gateways += @{
                Address     = $_.NextHop
                RouteMetric = $_.RouteMetric
            }
        }
    }

    $ipv4DNServerAddresses = Get-DNSClientServerAddress -InterfaceIndex $networkAdapter.InterfaceIndex -AddressFamily 'IPv4' -ErrorAction 'Ignore'
    if ( $ipv4DNServerAddresses ) {
        $ipv4DNServerAddresses.ServerAddresses | foreach {
            $naProperties.DNServers += $_
        }
    }

    $ipv6DNServerAddresses = Get-DNSClientServerAddress -InterfaceIndex $networkAdapter.InterfaceIndex -AddressFamily 'IPv6' -ErrorAction 'Ignore'
    if ( $ipv6DNServerAddresses ) {
        $ipv6DNServerAddresses.ServerAddresses | foreach {
            $naProperties.DNServers += $_
        }
    }

    Write-Output $( ConvertTo-Json -InputObject $naProperties )
`)

//------------------------------------------------------------------------------
