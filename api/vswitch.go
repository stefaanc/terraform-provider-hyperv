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
    Name                           string   // required
    SwitchType                     string   // "private" (default), "internal" or "external" - any other value is treated as "external"
    Notes                          string

    // external network adapters            // required when SwitchType is "external" - best not to specify any when not "external"
    //     specify NetworkAdapterName to use an existing adapter
    //     specify NetAdapterInterfaceDescription to disable the existing adapter and create new adapter with same name as switch
    //     if NetAdapterName is specified, it overrides NetAdapterInterfaceDescription
    AllowManagementOS              bool
    NetAdapterName                 string
    NetAdapterInterfaceDescription string
}

//------------------------------------------------------------------------------

func (c *HypervClient) CreateVSwitch(vsProperties *VSwitch) error {
    if vsProperties.Name == "" {
        return fmt.Errorf("[ERROR][terraform-provider-hyperv/api/CreateVSwitch(vsProperties)] missing 'vsProperties.Name'")
    }
    if vsProperties.SwitchType == "" {
        return fmt.Errorf("[ERROR][terraform-provider-hyperv/api/CreateVSwitch(vsProperties)] missing 'vsProperties.SwitchType'")
    }
    if strings.ToLower(vsProperties.SwitchType) != "private" && strings.ToLower(vsProperties.SwitchType) != "internal" && vsProperties.NetAdapterName == "" && vsProperties.NetAdapterInterfaceDescription == "" {
        return fmt.Errorf("[ERROR][terraform-provider-hyperv/api/CreateVSwitch(vsProperties)] missing 'vsProperties.NetAdapterName' or 'vsProperties.NetAdapterInterfaceDescription' for \"external\" switch")
    }

    return createVSwitch(c, vsProperties)
}

func (c *HypervClient) ReadVSwitch(vs *VSwitch) (vswitch *VSwitch, err error) {
    if vs.Name == "" {
        return nil, fmt.Errorf("[ERROR][terraform-provider-hyperv/api/vs.Read()] missing 'vs.Name'")
    }

    return readVSwitch(c, vs)
}

func (c *HypervClient) UpdateVSwitch(vs *VSwitch, vsProperties *VSwitch) error {
    if vs.Name == "" {
        return fmt.Errorf("[ERROR][terraform-provider-hyperv/api/vs.Update(vsProperties)] missing 'vs.Name'")
    }

    if vsProperties.SwitchType == "" {
        return fmt.Errorf("[ERROR][terraform-provider-hyperv/api/vs.Update(vsProperties)] missing 'vsProperties.SwitchType'")
    }
    if strings.ToLower(vsProperties.SwitchType) == "private" && strings.ToLower(vsProperties.SwitchType) != "internal" && vsProperties.NetAdapterName == "" && vsProperties.NetAdapterInterfaceDescription == "" {
        return fmt.Errorf("[ERROR][terraform-provider-hyperv/api/vs.Update(vsProperties)] missing 'vsProperties.NetAdapterName' or 'vsProperties.NetAdapterInterfaceDescription' for \"external\" switch")
    }

    return updateVSwitch(c, vs, vsProperties)
}

func (c *HypervClient) DeleteVSwitch(vs *VSwitch) error {
    if vs.Name == "" {
        return fmt.Errorf("[ERROR][terraform-provider-hyperv/api/vs.Delete()] missing 'vs.Name'")
    }

    return deleteVSwitch(c, vs)
}

//------------------------------------------------------------------------------

func createVSwitch(c *HypervClient, vsProperties *VSwitch) error {
    // convert vsProperties to JSON
    vsPropertiesJSON, err := json.Marshal(vsProperties)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/api/createVSwitch()] cannot cannot convert 'vsProperties' to json for %q\n", vsProperties.Name)
        return err
    }

    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, createVSwitchScript, createVSwitchArguments{
        VSPropertiesJSON: string(vsPropertiesJSON),
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-hyperv/api/createVSwitch()] cannot create vswitch %q\n", vsProperties.Name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/createVSwitch()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-hyperv/api/createVSwitch()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-hyperv/api/createVSwitch()] script stderr: %s", stderr.String())
        return err
    }

    log.Printf("[INFO][terraform-provider-hyperv/api/createVSwitch()] created vswitch %q\n", vsProperties.Name)
    return nil
}

type createVSwitchArguments struct{
    VSPropertiesJSON string
}

var createVSwitchScript = script.New("createVSwitch", "powershell", `
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

$vsProperties = $( ConvertFrom-Json -InputObject '{{.VSPropertiesJSON}}' )

$VMSwitchObject = Get-VMSwitch -Name $vsProperties.Name -ErrorAction 'SilentlyContinue'
if ($VMSwitchObject) {
    throw "[ERROR][terraform-provider-hyperv/api/createVSwitch()] vswitch '$($vsProperties.Name)' already exists"
}

$arguments = @{
    Name  = $vsProperties.Name
    Notes = $vsProperties.Notes
}

if ( $vsProperties -and $vsProperties.SwitchType -and ($vsProperties.SwitchType.ToLower() -eq "private") -or ($vsProperties.SwitchType.ToLower() -eq "internal") ) {
    $arguments.SwitchType = [Microsoft.HyperV.PowerShell.VMSwitchType]$vsProperties.SwitchType
} else {
    $arguments.AllowManagementOS = $vsProperties.AllowManagementOS
    if ($vsProperties.NetAdapterName) {
        $arguments.NetAdapterName = $vsProperties.NetAdapterName
    } else {
        $arguments.NetAdapterInterfaceDescription = $vsProperties.NetAdapterInterfaceDescription
    }
}

New-VMSwitch @arguments | Out-Default
`)

//------------------------------------------------------------------------------

func readVSwitch(c *HypervClient, vs *VSwitch) (vswitch *VSwitch, err error) {
    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err = runner.Run(c, readVSwitchScript, readVSwitchArguments{
        Name: vs.Name,
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] cannot read vswitch %q\n", vs.Name)
log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] script command: %s", runnerErr.Command())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] script stderr: %s", stderr.String())
        return nil, err
    }

    // convert stdout-JSON to vswitch
    vswitch = new(VSwitch)
    err = json.Unmarshal(stdout.Bytes(), vswitch)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] cannot convert json to 'vswitch' for %q\n", vs.Name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/readVSwitch()] json: %s", stdout.String())
        return nil, err
    }

    log.Printf("[INFO][terraform-provider-hyperv/api/readVSwitch()] read vswitch %q\n", vs.Name)
    return vswitch, nil
}

type readVSwitchArguments struct{
    Name string
}

var readVSwitchScript = script.New("readVSwitch", "powershell", `
$ErrorActionPreference = 'Stop'

$VMSwitchObject = Get-VMSwitch -Name '{{.Name}}'
if (-not $VMSwitchObject) {
    throw "[ERROR][terraform-provider-hyperv/api/createVSwitch()] cannot find vswitch '{{.Name}}'"
}

$VSwitch = @{
    Name              = $VMSwitchObject.Name
    SwitchType        = ([string]$VMSwitchObject.SwitchType).ToLower()
    Notes             = $VMSwitchObject.Notes
    AllowManagementOS = $VMSwitchObject.AllowManagementOS
}

if ($VMSwitchObject.NetAdapterInterfaceDescription) {
    $VSwitch.NetAdapterName                 = $( Get-NetAdapter -InterfaceDescription $VMSwitchObject.NetAdapterInterfaceDescription ).Name
    $VSwitch.NetAdapterInterfaceDescription = $VMSwitchObject.NetAdapterInterfaceDescription
}

Write-Output $(ConvertTo-Json -InputObject $VSwitch)
`)

//------------------------------------------------------------------------------

func updateVSwitch(c *HypervClient, vs *VSwitch, vsProperties *VSwitch) error {
    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // convert vsProperties to JSON
    vsPropertiesJSON, err := json.Marshal(vsProperties)
    if err != nil {
        return err
    }

    // run script
    err = runner.Run(c, updateVSwitchScript, updateVSwitchArguments{
        Name:             vs.Name,
        VSPropertiesJSON: string(vsPropertiesJSON),
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-hyperv/api/updateVSwitch()] cannot update vswitch %q\n", vs.Name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/updateVSwitch()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-hyperv/api/updateVSwitch()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-hyperv/api/updateVSwitch()] script stderr: %s", stderr.String())
        return err
    }

    log.Printf("[INFO][terraform-provider-hyperv/api/updateVSwitch()] updated vswitch %q\n", vs.Name)
    return nil
}

type updateVSwitchArguments struct{
    Name             string
    VSPropertiesJSON string
}

var updateVSwitchScript = script.New("updateVSwitch", "powershell", `
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

$VMSwitchObject = Get-VMSwitch -Name '{{.Name}}'
if (-not $VMSwitchObject) {
    throw "[ERROR][terraform-provider-hyperv/api/updateVSwitch()] cannot find vswitch '{{.Name}}'"
}

$vsProperties = $( ConvertFrom-Json -InputObject '{{.VSPropertiesJSON}}' )

$arguments = @{
    VMSwitch = $VMSwitchObject
    Notes    = $vsProperties.Notes
}

if ( $vsProperties -and $vsProperties.SwitchType -and (($vsProperties.SwitchType.ToLower() -eq "private") -or ($vsProperties.SwitchType.ToLower() -eq "internal")) ) {
    $arguments.SwitchType = [Microsoft.HyperV.PowerShell.VMSwitchType]$vsProperties.SwitchType
} else {
    $arguments.AllowManagementOS = $vsProperties.AllowManagementOS
    if ($vsProperties.NetAdapterName) {
        $arguments.NetAdapterName = $vsProperties.NetAdapterName
    } else {
        $arguments.NetAdapterInterfaceDescription = $vsProperties.NetAdapterInterfaceDescription
    }
}

Set-VMSwitch @arguments | Out-Default
`)

//------------------------------------------------------------------------------

func deleteVSwitch(c *HypervClient, vs *VSwitch) error {
    // create buffer to capture stdout & stderr
    var stdout bytes.Buffer
    var stderr bytes.Buffer

    // run script
    err := runner.Run(c, deleteVSwitchScript, deleteVSwitchArguments{
        Name: vs.Name,
    }, &stdout, &stderr)
    if err != nil {
        var runnerErr runner.Error
        errors.As(err, &runnerErr)
        log.Printf("[ERROR][terraform-provider-hyperv/api/deleteVSwitch()] cannot delete vswitch %q\n", vs.Name)
        log.Printf("[ERROR][terraform-provider-hyperv/api/deleteVSwitch()] script exitcode: %d", runnerErr.ExitCode())
        log.Printf("[ERROR][terraform-provider-hyperv/api/deleteVSwitch()] script stdout: %s", stdout.String())
        log.Printf("[ERROR][terraform-provider-hyperv/api/deleteVSwitch()] script stderr: %s", stderr.String())
        return err
    }

    log.Printf("[INFO][terraform-provider-hyperv/api/deleteVSwitch()] deleted vswitch %q\n", vs.Name)
    return nil
}

type deleteVSwitchArguments struct{
    Name string
}

var deleteVSwitchScript = script.New("deleteVSwitch", "powershell", `
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'   # progress-bar fails when using ssh

$VMSwitchObject = Get-VMSwitch -Name '{{.Name}}'
if (-not $VMSwitchObject) {
    throw "[ERROR][terraform-provider-hyperv/api/deleteVSwitch()] cannot find vswitch '{{.Name}}'"
}

Remove-VMSwitch -VMSwitch $VMSwitchObject -Force | Out-Default
`)

//------------------------------------------------------------------------------
