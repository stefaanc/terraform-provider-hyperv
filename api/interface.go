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

type Interface struct {
    Index               uint32
    Name                string
    Alias               string
    Description         string

    MACAddress          string
    NetworkAdapterName  string
    VNetworkAdapterName string

    NetworkName         string
    ComputerName        string
}

//------------------------------------------------------------------------------

func (c *HypervClient) ReadInterface(iQuery *Interface) (iProperties *Interface, err error) {
    if iQuery.Index == 0 &&
       iQuery.Name == "" &&
       iQuery.Alias == "" &&
       iQuery.Description == "" &&
       iQuery.MACAddress == "" &&
       iQuery.NetworkAdapterName == "" &&
       iQuery.VNetworkAdapterName == "" {
        return nil, fmt.Errorf("[ERROR][terraform-provider-hyperv/api/ReadInterface(iQuery)] empty 'iQuery'")
    }

    return readInterface(c, iQuery)
}

//------------------------------------------------------------------------------

func readInterface(c *HypervClient, iQuery *Interface) (iProperties *Interface, err error) {
    // find id in iQuery
    var id interface{}
    if        iQuery.Index != 0                { id = iQuery.Index
    } else if iQuery.Name != ""                { id = iQuery.Name
    } else if iQuery.Alias != ""               { id = iQuery.Alias
    } else if iQuery.Description != ""         { id = iQuery.Description
    } else if iQuery.MACAddress != ""          { id = iQuery.MACAddress
    } else if iQuery.NetworkAdapterName != ""  { id = iQuery.NetworkAdapterName
    } else if iQuery.VNetworkAdapterName != "" { id = iQuery.VNetworkAdapterName
    }

    // convert iQuery to JSON
    iQueryJSON, err := json.Marshal(iQuery)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/api/readInterface()] cannot cannot convert 'iQuery' to json for interface %#v\n", id)
        return nil, err
    }

    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, readInterfaceScript, readInterfaceArguments{
        IQueryJSON: string(iQueryJSON),
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readInterface()] cannot read interface %#v\n", id)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readInterface()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readInterface()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readInterface()] script stderr: %s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-hyperv/api/readInterface()] runner: %s", stderr.String())
        }

        return nil, err
    }

    // convert stdout-JSON to iProperties
    iProperties = new(Interface)
    err = json.Unmarshal(stdout.Bytes(), iProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/api/readInterface()] cannot convert json to 'iProperties' for interface %#v\n", id)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readInterface()] json: %s", stdout.String())
        return nil, err
    }

    log.Printf("[INFO][terraform-provider-hyperv/api/readInterface()] read interface %#v\n", id)
    return iProperties, nil
}

type readInterfaceArguments struct{
    IQueryJSON string
}

var readInterfaceScript = script.New("readInterface", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    # convert vsProperties to JSON
    $iQuery = $( ConvertFrom-Json -InputObject '{{.IQueryJSON}}' )

    # find network-adapter
    if ( $iQuery.Index -ne 0 ) {
        $id = $iQuery.Index
        $networkAdapter = Get-NetAdapter -InterfaceIndex $id -ErrorAction 'Ignore'
    }
    elseif ( $iQuery.Name -ne "" ) {
        $id = $iQuery.Name
        Get-NetAdapter -IncludeHidden | foreach {
            if ( $iQuery.Name -eq $_.InterfaceName ) {
                $networkAdapter = $_
            }
        }
    }
    elseif ( $iQuery.Alias -ne "" ) {
        $id = $iQuery.Alias
        Get-NetAdapter -IncludeHidden | foreach {
            if ( $iQuery.Alias -eq $_.InterfaceAlias ) {
                $networkAdapter = $_
            }
        }
    }
    elseif ( $iQuery.Description -ne "" ) {
        $id = $iQuery.Description
        $networkAdapter = Get-NetAdapter -InterfaceDescription $id -ErrorAction 'Ignore'
    }
    elseif ( $iQuery.MACAddress -ne "" ) {
        $id = $iQuery.MACAddress
        Get-NetAdapter -IncludeHidden | foreach {
            if ( $iQuery.MACAddress -eq $_.MacAddress ) {
                $networkAdapter = $_
            }
        }
    }
    elseif ( $iQuery.NetworkAdapterName -ne "" ) {
        $id = $iQuery.NetworkAdapterName
        $networkAdapter = Get-NetAdapter -Name $id -ErrorAction 'Ignore'
    }
    elseif ( $iQuery.VNetworkAdapterName -ne "" ) {
        $id = $iQuery.VNetworkAdapterName
        $vnetworkAdapter = Get-VMNetworkAdapter -All -Name $id -ErrorAction 'Ignore'
        if ( $vnetworkAdapter ) {
            $vnetworkAdapterName = $vnetworkAdapter.Name
            $networkAdapterName = "vEthernet (${vnetworkAdapterName})"
            Get-NetAdapter -IncludeHidden | foreach {
                # try matching MAC-address or name
                if ( ( $vnetworkAdapter.MacAddress -eq $_.MacAddress ) -or ( $networkAdapterName -eq $_.Name ) ) {
                    $networkAdapter = $_
                }
            }
        }
    }
    if ( -not $networkAdapter ) {
        throw "cannot find interface '$id'"
    }

    # find vnetwork-adapter
    if ( ( -not $vnetworkAdapter ) -and ( $networkAdapter.DriverDescription -eq "Hyper-V Virtual Ethernet Adapter" ) ) {
        if ( $networkAdapter.Name -match "vEthernet \((?<name>.*)\)" ) {
            $vnetworkAdapterName = $matches["name"]
        }
        Get-VMNetworkAdapter -All | foreach {
            # try matching MAC-address or name
            if ( ( $networkAdapter.MacAddress -eq $_.MacAddress ) -or ( $vnetworkAdapterName -eq $_.Name ) ) {
                $vnetworkAdapter = $_
                $vnetworkAdapterName = $_.Name
            }
        }
    }

    # prepare result
    $iProperties = @{
        Index               = $networkAdapter.InterfaceIndex
        Name                = $networkAdapter.InterfaceName
        Alias               = $networkAdapter.InterfaceAlias
        Description         = $networkAdapter.InterfaceDescription
        MACAddress          = $networkAdapter.MacAddress
        NetworkAdapterName  = $networkAdapter.Name
        VNetworkAdapterName = $vnetworkAdapterName
        NetworkName         = $( Get-NetConnectionProfile -InterfaceIndex $networkAdapter.InterfaceIndex ).Name
        ComputerName        = $networkAdapter.SystemName
    }

    Write-Output $( ConvertTo-Json -InputObject $iProperties )
`)

//------------------------------------------------------------------------------
