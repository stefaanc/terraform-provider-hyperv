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

type VSwitch struct {
    Name            string
    NetworkAdapters []string
    Notes           string
}

//------------------------------------------------------------------------------

func (c *HypervClient) ReadVSwitch(vsQuery *VSwitch) (vsProperties *VSwitch, err error) {
    if vsQuery.Name == "" {
        return nil, fmt.Errorf("[ERROR][terraform-provider-hyperv/api/ReadVSwitch(vsQuery)] missing 'vsQuery.Name'")
    }

    return readVSwitch(c, vsQuery)
}

//------------------------------------------------------------------------------

func readVSwitch(c *HypervClient, vsQuery *VSwitch) (vsProperties *VSwitch, err error) {
    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, readVSwitchScript, readVSwitchArguments{
        Name: vsQuery.Name,
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] cannot read vswitch %q\n", vsQuery.Name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] script stderr: %s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-hyperv/api/readVSwitch()] runner: %s", stderr.String())
        }

        return nil, err
    }

    // convert stdout-JSON to vsProperties
    vsProperties = new(VSwitch)
    err = json.Unmarshal(stdout.Bytes(), vsProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] cannot convert json to 'vsProperties' for vswitch %q\n", vsQuery.Name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] json: %s", stdout.String())
        return nil, err
    }

    log.Printf("[INFO][terraform-provider-hyperv/api/readVSwitch()] read vswitch %q\n", vsQuery.Name)
    return vsProperties, nil
}

type readVSwitchArguments struct{
    Name string
}

var readVSwitchScript = script.New("readVSwitch", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $vmswitch = Get-VMSwitch -Name '{{.Name}}' -ErrorAction 'Ignore'
    if ( -not $vmswitch ) {
        throw "cannot find vswitch '{{.Name}}'"
    }

    $vsProperties = @{
        Name            = $vmswitch.Name
        NetworkAdapters = @()
        Notes           = $vmswitch.Notes
    }

    if ( $vmswitch.NetAdapterInterfaceDescriptions ) {
        $vmswitch.NetAdapterInterfaceDescriptions | foreach {
            $vsProperties.NetworkAdapters += $( Get-NetAdapter -InterfaceDescription $_ ).Name
        }
    }

    Write-Output $( ConvertTo-Json -InputObject $vsProperties )
`)

//------------------------------------------------------------------------------
