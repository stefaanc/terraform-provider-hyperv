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

type Network struct {
    Name              string
    ConnectionProfile string
}

//------------------------------------------------------------------------------

func (c *HypervClient) ReadNetwork(nQuery *Network) (nProperties *Network, err error) {
    if nQuery.Name == "" {
        return nil, fmt.Errorf("[ERROR][terraform-provider-hyperv/api/ReadNetwork(nQuery)] missing 'nQuery.Name'")
    }

    return readNetwork(c, nQuery)
}

//------------------------------------------------------------------------------

func readNetwork(c *HypervClient, nQuery *Network) (nProperties *Network, err error) {
    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, readNetworkScript, readNetworkArguments{
        Name: nQuery.Name,
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetwork()] cannot read network %#v\n", nQuery.Name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetwork()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetwork()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetwork()] script stderr: %s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-hyperv/api/readNetwork()] runner: %s", stderr.String())
        }

        return nil, err
    }

    // convert stdout-JSON to nProperties
    nProperties = new(Network)
    err = json.Unmarshal(stdout.Bytes(), nProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetwork()] cannot convert json to 'nProperties' for network %#v\n", nQuery.Name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readNetwork()] json: %s", stdout.String())
        return nil, err
    }

    log.Printf("[INFO][terraform-provider-hyperv/api/readNetwork()] read network %#v\n", nQuery.Name)
    return nProperties, nil
}

type readNetworkArguments struct{
    Name string
}

var readNetworkScript = script.New("readNetwork", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $networkConnectionProfile = Get-NetConnectionProfile -Name '{{.Name}}' -ErrorAction 'Ignore'
    if ( -not $networkConnectionProfile ) {
        throw "cannot find network '{{.Name}}'"
    }

    # prepare result
    $nProperties = @{
        Name              = $networkConnectionProfile.Name
        ConnectionProfile = $networkConnectionProfile.NetworkCategory.ToString()
    }

    Write-Output $( ConvertTo-Json -InputObject $nProperties )
`)

//------------------------------------------------------------------------------
