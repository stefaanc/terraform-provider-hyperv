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

type VNetworkAdapter struct {
    Name                    string   // remark that there can be multiple vnetwork_adapters with the same name (for different vmachines/management_os)
    VMachineName            string   // remark that a vmachine/management_os can have multiple vnetwork_adapters (with different names)

    MACAddress              string   // when dynamic (as opposed to static ), has a value only when connected to a vswitch
    AllowMACAddressSpoofing bool

    VSwitchName             string   // has a value only when connected to a vswitch
}

//------------------------------------------------------------------------------

func (c *HypervClient) ReadVNetworkAdapter(vnaQuery *VNetworkAdapter) (vnaProperties *VNetworkAdapter, err error) {
    return readVNetworkAdapter(c, vnaQuery)
}

//------------------------------------------------------------------------------

func readVNetworkAdapter(c *HypervClient, vnaQuery *VNetworkAdapter) (vnaProperties *VNetworkAdapter, err error) {
    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    full_name := vnaQuery.Name
    if vnaQuery.VMachineName != "" {
        full_name += "@" + vnaQuery.VMachineName
    }

    // run script
    err = runner.Run(c, readVNetworkAdapterScript, readVNetworkAdapterArguments{
        Name:         vnaQuery.Name,
        VMachineName: vnaQuery.VMachineName,
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVNetworkAdapter()] cannot read network_adapter %#v\n", full_name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVNetworkAdapter()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVNetworkAdapter()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVNetworkAdapter()] script stderr: %s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-hyperv/api/readVNetworkAdapter()] runner: %s", stderr.String())
        }

        return nil, err
    }

    // convert stdout-JSON to vnaProperties
    vnaProperties = new(VNetworkAdapter)
    err = json.Unmarshal(stdout.Bytes(), vnaProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVNetworkAdapter()] cannot convert json to 'vnaProperties' for vnetwork_adapter %#v\n", full_name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVNetworkAdapter()] json: %s", stdout.String())
        return nil, err
    }

    log.Printf("[INFO][terraform-provider-hyperv/api/readVNetworkAdapter()] read vnetwork_adapter %#v\n", full_name)
    return vnaProperties, nil
}

type readVNetworkAdapterArguments struct{
    Name         string
    VMachineName string
}

var readVNetworkAdapterScript = script.New("readVNetworkAdapter", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $Name         = '{{.Name}}'
    $VMachineName = '{{.VMachineName}}'

    if ( $VMachineName -and $Name ) {
        $vnetworkAdapter = Get-VMNetworkAdapter -Name $Name -VMName $VMachineName -ErrorAction 'Ignore'
    }
    elseif ( $VMachineName ) {
        $vnetworkAdapter = Get-VMNetworkAdapter -VMName $VMachineName -ErrorAction 'Ignore'
        if ( $vnetworkAdapter -and ( $vnetworkAdapter.length -gt 1 ) ) { $vnetworkAdapter = $null }
    }
    elseif ( $Name ) {
        $vnetworkAdapter = Get-VMNetworkAdapter -Name $Name -ManagementOS -ErrorAction 'Ignore'

        if ( -not $vnetworkAdapter ) {
            $vnetworkAdapter = Get-VMNetworkAdapter -Name $Name -All -ErrorAction 'Ignore'
            if ( $vnetworkAdapter -and ( $vnetworkAdapter.length -gt 1 ) ) { $vnetworkAdapter = $null }
        }
    }
    if ( -not $vnetworkAdapter ) {
        $fullName = $Name
        if ( $VMachineName -ne "" ) { $fullName += "@" + $VMachineName }

        throw "cannot find vnetwork_adapter '$fullName'"
    }

    # prepare result
    $vnaProperties = @{
        Name                    = $vnetworkAdapter.Name
        VMachineName            = if ( -not $vnetworkAdapter.IsManagementOs ) { $vnetworkAdapter.VMName } else { "" }

        MACAddress              = ( $vnetworkAdapter.MacAddress -split '(\w{2})' | where { $_ } ) -join "-"
        AllowMACAddressSpoofing = ( $vnetworkAdapter.MacAddressSpoofing -eq "On" )

        VSwitchName             = $vnetworkAdapter.SwitchName
    }

    Write-Output $( ConvertTo-Json -InputObject $vnaProperties )
`)

//------------------------------------------------------------------------------
