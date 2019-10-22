# Terraform Provider Hyper-V

**a terraform provider to work with hyper-V**



<br/>

### !!! UNDER CONSTRUCTION !!!!!!!!



<br/>

## Prerequisites

To build:
- [GNU make](https://www.gnu.org/software/make/)
- [Golang](https://golang.org/) >= v1.13
- [Terraform plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk) ~= v1.0.0

To use:
- [Terraform](https://terraform.io) >= v0.12.9



<br>

## Building The Provider

1. Clone the git-repository on your machine

   ```shell
   mkdir -p $my_repositories
   cd $my_repositories
   git clone git@github.com:stefaanc/terraform-provider-hyperv
   ```
   > `$my_repositories` must point to the directory where you want to clone the repository
   
2. Build the provider

   ```shell
   cd $my_repositories/terraform-provider-hosts
   make release
   ```

   This will build the provider and put it in 
   - `%AppData%\terraform.d\plugins` on Windows
   - `$HOME\.terraform.d\plugins` on Linux
<br/>

 > :bulb:  
 > The makefile provides more commands: `tidy`, `test`, `log`, `report`, `testacc`, `build`, ...
    


<br>

## Installing The Provider

1. Download the provider to your machine

   - go to [the releases tab on github](https://github.com/stefaanc/terraform-provider-hyperv/releases)
   - download the file that is appropriate for your machine

2. Move the provider from your `Downloads` folder to

   - `%AppData%\terraform.d\plugins` on Windows
   - `$HOME\.terraform.d\plugins` on Linux
<br/>

> :bulb:  
> Alternatively, you can try our latest release-in-progress under the `releases` folder.  No guarantee though this will be a fully working provider.



<br>

## Using The Provider

> :bulb:  
> You can find the some of following examples (and more) under the `examples` folder

### provider "hyperv"

Configures a provider for a Hyper-V server.
 
```terraform
provider "hyperv" {}
```
 
```terraform
provider "hyperv" {
    type = "local"
}
```

```terraform
provider "hyperv" {
    type     = "ssh"
    host     = "localhost"
    port     = 22
    user     = "me"
    password = "my-password"
    insecure = true
}
```

Arguments  | &nbsp;   | Description
:----------|:--------:|:-----------
`type`     | Optional | The type of connection to the hyperv-server: `"local"` or `"ssh"`.  <br/>- defaults to `"local"`
` `        | ` `      | ` `
`host`     | Optional | The hyperv-server. <br/>- ignored when `type = "local"` <br/>- defaults to `"localhost"`
`port`     | Optional | The hyperv-server's port for ssh. <br/>- ignored when `type = "local"` <br/>- defaults to `22`
`user`     | Optional | The user name for communication with the hyperv-server. <br/>- ignored when `type = "local"` <br/>- required when `type = "ssh"`
`password` | Optional | The user password for communication with the hyperv-server. <br/>- ignored when `type = "local"` <br/>- required when `type = "ssh"`
`insecure` | Optional | Allow insecure communication - disables checking of the server certificate. <br/>- ignored when `type = "local"` <br/>- defaults to `false` <br/><br/> When `insecure = false`, the hyperv-server's certificate is checked against the user's known hosts, as specified by the file `~/.ssh/known_hosts`.  

> :bulb:  
> The Hyper-V API needs elevated credentials ("Run as Administrator") for all methods.
> When using `type = "local"`, you need to run terraform from an elevated shell.
> When using `type = "ssh"`, terraform will always use the most elevated credentials available to the configured user.



<br>

## Data-sources

### data "hyperv_vswitch"

Reads a Hyper-V virtual switch.

```terraform
data "hyperv_vswitch" "default" {
    name = "Default Switch"
}
```

Arguments     | &nbsp;   | Description
:-------------|:--------:|:-----------
`name`        | Required | The name of the virtual switch.
` `           | ` `      | ` `
`x_lifecycle` | Optional | see [x_lifecycle for data-sources](#extended-lifecycle-customizations-for-data-sources)
  
Exports                             | &nbsp;   | Description
:-----------------------------------|:--------:|:-----------
`switch_type`                       | Computed | The type of virtual switch: `"private"`, `"internal"` or `"external"`.
`notes`                             | Computed | Notes added to the virtual switch.
`allow_management_os`               | Computed | The hyperv-server is allowed to participate into the communication on the virtual switch. 
`net_adapter_name`                  | Computed | The name of the network adapter used for an "external" virtual switch.
`net_adapter_interface_description` | Computed | The description for the network adapter interface used for an "external" virtual switch.



<br>

### extended lifecycle customizations for data-sources

The `x_lifecycle` block defines extensions to the terraform lifecycle customizations meta-data for data-sources (although at moment of writing this text, there are no such lifecycle customizations defined for data-sources).  As opposed to the terraform meta-data, that can be added to all of the configured data-sources, the `x_lifecycle` block can only be added to the data-sources that implement this block.

```terraform
data "hyperv_vswitch" "default" {
    name = "Default Switch"

    x_lifecycle {
        ignore_error_if_not_exists = true
    }
}
```

Arguments                                | &nbsp;   | Description
:----------------------------------------|:--------:|:-----------
`x_lifecycle.ignore_error_if_not_exists` | Optional | Ignores the "cannot find" or "doesn't exist" errors from the API.  <br/><br/>This can be used to test if a data-source exists.  For example, the "Default Switch" doesn't exist in older versions of Hyper-V, and does exist by default in newer versions of Hyper-V.
  
Exports              | &nbsp;   | Description
:--------------------|:--------:|:-----------
`x_lifecycle.exists` | Computed | Set to `true` if the data-source exists.


<br>

## Resources

### resource "hyperv_vswitch"   

```terraform
resource "hyperv_vswitch" "private" {
    provider = hyperv.local

    name                = "Private Switch"
    switch_type         = "private"
    notes               = "private notes"
}
```

```terraform
resource "hyperv_vswitch" "internal" {
    provider = hyperv.local

    name                = "Internal Switch"
    switch_type         = "internal"
    notes               = "internal notes"
}
```

```terraform
resource "hyperv_vswitch" "external" {
    provider = hyperv.local

    name                = "External Switch"
    switch_type         = "external"
    notes               = "external notes"

    allow_management_os = true
    net_adapter_interface_description = "Intel(R) 82579LM Gigabit Network Connection"
}
```

Arguments                           | &nbsp;   | Description
:-----------------------------------|:--------:|:-----------
`name`                              | Required | The name of the virtual switch.
`switch_type`                       | Required | The type of virtual switch: `"private"`, `"internal"` or `"external"`.
`notes`                             | Optional | Notes added to the virtual switch.
` `                                 | ` `      | ` `
`allow_management_os`               | Optional | The hyperv-server is allowed to participate into the communication on the virtual switch.  <br/>- must not be configured or set to `false` when `switch_type = "private"`.  <br/>- must not be configured or set to `true` when `switch_type = "internal"`  <br/>- defaults to `false` when `switch_type = "external"`
`net_adapter_name`                  | Optional | Use the existing network adapter with this name.  <br/>- must not be configured when `switch_type = "private"` or `switch_type = "internal"`  <br/>- must not be configured  when `switch_type = "external"` and `net_adapter_interface_description` is configured  <br/>- required when `switch_type = "external"` and `net_adapter_interface_description` is not configured 
`net_adapter_interface_description` | Optional | Disable existing network adapter and create new network adapter for this interface.  <br/>- must not be configured when `switch_type = "private"` or `switch_type = "internal"`  <br/>- must not be configured when `switch_type = "external"` and `net_adapter_name` is configured  <br/>- required when `switch_type = "external"` and `net_adapter_name` is not configured
` `                                 | ` `      | ` `
`x_lifecycle`                       | Optional | see [x_lifecycle for resources](#extended-lifecycle-customizations-for-resources)
  
Exports                             | &nbsp;   | Description
:-----------------------------------|:--------:|:-----------
`allow_management_os`               | Computed | The hyperv-server is allowed to participate into the communication on the virtual switch. 
`net_adapter_name`                  | Computed | The name of the network adapter used for an "external" virtual switch.
`net_adapter_interface_description` | Computed | The description for the network adapter interface used for an "external" virtual switch.

**_Importing a hyperv_vswitch using terraform import_**

You can import a virtual switch using the switch's name as an import ID.

- Assuming a configuration

  ```terraform
  provider "hyperv" {}

  resource "hyperv_vswitch" "default" {
      name        = "Default Switch"
      switch_type = "internal"
      notes       = "internal notes"
  }
  ```

  When terraform tries to create this then this will fail when the virtual switch already exists.  You can delete the switch from the infrastructure, and then re-create it using terraform.  However, this may be a bit more involved when you need to automate this.  Alternatively you can import it.

- Run the terraform import command using the name of the switch as import ID.

  ```shell
  terraform import "hyperv_vswitch.default" "Default Switch"
  ```

  The resource will be imported into the terraform state, and the usual lifecycle will be applied next time `terraform apply` is run.



<br>

### extended lifecycle customizations for resources

The `x_lifecycle` block defines extensions to the terraform lifecycle customizations meta-data for resources.  As opposed to the terraform meta-data, that can be added to all of the configured resources, the `x_lifecycle` block can only be added to the resources that implement this block.

```terraform
resource "hyperv_vswitch" "default" {
    name        = "Default Switch"
    switch_type = "internal"

    x_lifecycle {
        import_if_exists = true
    }
}
```

Arguments                         | &nbsp;   | Description
:---------------------------------|:--------:|:-----------
`x_lifecycle.import_if_exists`    | Optional | Imports the resource when it does exist, avoiding the "already exists" errors from the API.  <br/><br/>This can be used in cases where existence of a resource is unknown and would require "obscure" configuration to test and decide if the resource needs creating.  For example, the "Default Switch" doesn't exist in older versions of Hyper-V, and does exist by default in newer versions of Hyper-V.
`x_lifecycle.destroy_if_imported` | Optional | Destroys the imported resource when using `terraform destroy`.  <br/><br/>By default, a resource that is imported using `import_if_exists = "true"` is **not** destroyed when using `terraform destroy`.
  
Exports                | &nbsp;   | Description
:----------------------|:--------:|:-----------
`x_lifecycle.imported` | Computed | Set to `true` when the resource is imported using `import_if_exists = "true"`. 



<br>

## For Further Investigation

- add more/all arguments in-line with the PowerShell Hyper-V API
- add acceptance tests
- add API tests
- terraform-style documentation
