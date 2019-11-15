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

    $NetAdapterObject = Get-NetAdapter -Name '{{.Name}}' -ErrorAction 'Ignore'
    if ( -not $NetAdapterObject ) {
        throw "cannot find network_adapter '{{.Name}}'"
    }

    # prepare result
    $naProperties = @{
        Name              = $NetAdapterObject.Name
        MACAddress        = $NetAdapterObject.MacAddress

        IPAddresses       = @()
        Gateways          = @()
        DNServers         = @()

        AdminStatus       = $NetAdapterObject.AdminStatus.ToString()
        OperationalStatus = $NetAdapterObject.ifOperStatus.ToString()
        ConnectionStatus  = $NetAdapterObject.MediaConnectionState.ToString()
        ConnectionSpeed   = $NetAdapterObject.LinkSpeed.ToString()
        IsPhysical        = $NetAdapterObject.ConnectorPresent
    }

    $IPv4BindingObject = Get-NetAdapterBinding -Name '{{.Name}}' -ComponentID 'ms_tcpip' -ErrorAction 'Ignore'
    if ( $IPv4BindingObject ) {
        $naProperties.IPv4InterfaceDisabled = -not $IPv4BindingObject.Enabled
    }

    $IPv4InterfaceObject = Get-NetIPInterface -InterfaceIndex $NetAdapterObject.InterfaceIndex -AddressFamily 'IPv4' -ErrorAction 'Ignore'
    if ( $IPv4InterfaceObject ) {
        $naProperties.IPv4InterfaceMetric = $IPv4InterfaceObject.IPv4InterfaceMetric
    }

    $IPv6BindingObject = Get-NetAdapterBinding -Name '{{.Name}}' -ComponentID 'ms_tcpip6' -ErrorAction 'Ignore'
    if ( $IPv6BindingObject ) {
        $naProperties.IPv6InterfaceDisabled = -not $IPv6BindingObject.Enabled
    }

    $IPv6InterfaceObject = Get-NetIPInterface -InterfaceIndex $NetAdapterObject.InterfaceIndex -AddressFamily 'IPv6' -ErrorAction 'Ignore'
    if ( $IPv6InterfaceObject ) {
        $naProperties.IPv6InterfaceMetric = $IPv6InterfaceObject.IPv6InterfaceMetric
    }

    $DNSClientObject = Get-DNSClient -InterfaceIndex $NetAdapterObject.InterfaceIndex -ErrorAction 'Ignore'
    if ( $DNSClientObject ) {
        $naProperties.RegisterConnectionAddress = $DNSClientObject.RegisterThisConnectionsAddress
        if ( $DNSClientObject.UseSuffixWhenRegistering ) {
            $naProperties.RegisterConnectionSuffix  = $DNSClientObject.ConnectionSpecificSuffix
        }
    }

    $IPv4AddressObject = Get-NetIPAddress -InterfaceIndex $NetAdapterObject.InterfaceIndex -AddressFamily 'IPv4' -AddressState 'Preferred' -ErrorAction 'Ignore'
    if ( $IPv4AddressObject ) {
        $IPv4AddressObject | foreach {
            $naProperties.IPAddresses += @{
                Address      = $_.IPAddress
                PrefixLength = $_.PrefixLength
                SkipAsSource = $_.SkipAsSource
            }
        }
    }

    $IPv6AddressObject = Get-NetIPAddress -InterfaceIndex $NetAdapterObject.InterfaceIndex -AddressFamily 'IPv6' -AddressState 'Preferred' -ErrorAction 'Ignore'
    if ( $IPv6AddressObject ) {
        $IPv6AddressObject | foreach {
            $naProperties.IPAddresses += @{
                Address      = $_.IPAddress
                PrefixLength = $_.PrefixLength
                SkipAsSource = $_.SkipAsSource
            }
        }
    }

    $IPv4RouteObject = Get-NetRoute -InterfaceIndex $NetAdapterObject.InterfaceIndex -AddressFamily 'IPv4' -DestinationPrefix '0.0.0.0/0' -ErrorAction 'Ignore'
    if ( $IPv4RouteObject ) {
        $IPv4RouteObject | foreach {
            $naProperties.Gateways += @{
                Address     = $_.NextHop
                RouteMetric = $_.RouteMetric
            }
        }
    }

    $IPv6RouteObject = Get-NetRoute -InterfaceIndex $NetAdapterObject.InterfaceIndex -AddressFamily 'IPv6' -DestinationPrefix '::/0' -ErrorAction 'Ignore'
    if ( $IPv6RouteObject ) {
        $IPv6RouteObject | foreach {
            $naProperties.Gateways += @{
                Address     = $_.NextHop
                RouteMetric = $_.RouteMetric
            }
        }
    }

    $IPv4DNServerAddressesObject = Get-DNSClientServerAddress -InterfaceIndex $NetAdapterObject.InterfaceIndex -AddressFamily 'IPv4' -ErrorAction 'Ignore'
    if ( $IPv4DNServerAddressesObject ) {
        $IPv4DNServerAddressesObject.ServerAddresses | foreach {
            $naProperties.DNServers += $_
        }
    }

    $IPv6DNServerAddressesObject = Get-DNSClientServerAddress -InterfaceIndex $NetAdapterObject.InterfaceIndex -AddressFamily 'IPv6' -ErrorAction 'Ignore'
    if ( $IPv6DNServerAddressesObject ) {
        $IPv6DNServerAddressesObject.ServerAddresses | foreach {
            $naProperties.DNServers += $_
        }
    }

    Write-Output $( ConvertTo-Json -InputObject $naProperties )
`)

//------------------------------------------------------------------------------
