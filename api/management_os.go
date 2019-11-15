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

type ManagementOS struct {
    Name string

    DNS_SuffixSearchList []string
    DNS_EnableDevolution bool
    DNS_DevolutionLevel  int32
}

//------------------------------------------------------------------------------

func (c *HypervClient) ReadManagementOS() (osProperties *ManagementOS, err error) {
    return readManagementOS(c)
}

//------------------------------------------------------------------------------

func readManagementOS(c *HypervClient) (osProperties *ManagementOS, err error) {
    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, readManagementOSScript, nil, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readManagementOS()] cannot read management_os\n")
        log.Printf("[ERROR][terraform-provider-hyperv/api/readManagementOS()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readManagementOS()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readManagementOS()] script stderr: %s", stderr.String())

        // get to the cause of a "runner failed" error to display in terraform UI
        if strings.Contains(runnerErr.Error(), "runner failed") {
            err = fmt.Errorf("[terraform-provider-hyperv/api/readManagementOS()] runner: %s", stderr.String())
        }

        return nil, err
    }

    // convert stdout-JSON to osProperties
    osProperties = new(ManagementOS)
    err = json.Unmarshal(stdout.Bytes(), osProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/api/readManagementOS()] cannot convert json to 'osProperties' for management_os\n")
        log.Printf("[ERROR][terraform-provider-hyperv/api/readManagementOS()] json: %s", stdout.String())
        return nil, err
    }

    log.Printf("[INFO][terraform-provider-hyperv/api/readManagementOS()] read management_os\n")
    return osProperties, nil
}

var readManagementOSScript = script.New("readManagementOS", "powershell", `
    $ErrorActionPreference = 'Stop'
    $ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

    $settings = Get-DnsClientGlobalSetting

    # prepare result
    $osProperties = @{
        Name = $env:ComputerName

        DNS_SuffixSearchList = $settings.SuffixSearchList
        DNS_EnableDevolution = $settings.UseDevolution
        DNS_DevolutionLevel  = $settings.DevolutionLevel
    }

    Write-Output $( ConvertTo-Json -InputObject $osProperties )
`)

//------------------------------------------------------------------------------
