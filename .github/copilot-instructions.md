# iacconsole-cli AI Coding Agent Instructions

This document provides essential guidance for AI agents working on the `iacconsole-cli` codebase.

## Project Overview & Architecture

`iacconsole-cli` is an infrastructure layers configuration orchestrator that dynamically manages OpenTofu or Terraform. It provides infrastructure configuration definitions from outside the Terraform code, using either files or an Infrastructure Layers Configuration Management Database (CMDB) called IaCConsole-DB. Its core purpose is to enable the reuse of Terraform code across multiple environments by separating configuration from the infrastructure code itself.

The key concepts are:
- **units**: Reusable, generic Terraform/OpenTofu code modules located in a `units_path` (e.g., `examples/units/`). Each subdirectory within an organization folder (e.g., `demo-org/vpc`) is a "unit".
- **Inventory (Configuration Sources)**: Configuration data that can come from two sources:
  1. **File-based**: JSON files that define the configuration for different environments (`examples/inventory/`).
  2. **IaCConsole-DB**: External OpenAPI-based Configuration Management Database (CMDB).
- **Dimensions**: Key-value pairs (`-d key:value`) that select specific configuration objects (from either source) to apply to a unit.
- **Orchestration Flow**: `iacconsole-cli` dynamically creates a temporary directory, copies the selected `unit` code, generates `.tfvars` files from the specified dimension's inventory sources, and then executes the `tofu` or `terraform` command within that directory.

The main application logic is in Go (`cmd/` and `utils/` directories), while the infrastructure code it manages is HCL (Terraform) located in the `examples/` directory.

## Key Files and Directories

- `main.go`: Main entry point for the CLI application.
- `cmd/exec.go`: Implements the primary `exec` command, which is the main orchestration logic.
- `utils/`: Contains the core Go helper functions.
  - `preparetemp.go`: Logic for creating the temporary execution directory.
  - `generatevars.go`: Logic for processing inventory files and generating Terraform variables.
  - `dimensions.go`: Handles the parsing and management of dimension data.
- `examples/inventory/`: Contains the hierarchical configuration data (JSON files). This is the "database" for the infrastructure configurations.
- `examples/units/`: Contains the reusable Terraform modules ("units").
- `examples/.iacconsolerc`: An example of the user-level configuration file (usually located at `$HOME/.iacconsolerc`). It defines paths and backend settings.

## Developer Workflow

The primary workflow involves running the `exec` command to prepare and execute a Terraform/OpenTofu action.

**Get the `iacconsole-cli` binary:**
- Download pre-built binaries for Linux and macOS from the [releases page](https://github.com/alt-dima/iacconsole-cli/releases)
- Or build it yourself:
```bash
go build -o bin/iacconsole-cli .
```

**Run a typical command:**
This command initializes the `vpc` unit for the `test-account` in the `staging1` datacenter of the `demo-org`.
```bash
./bin/iacconsole-cli exec -o demo-org -d account:test-account -d datacenter:staging1 -u vpc -- init
```
- `-o`: Organization (subfolder in `inventory` and `units`).
- `-d`: Dimension (maps to a config file, e.g., `inventory/demo-org/datacenter/staging1.json`).
- `-t`: unit (the Terraform code to use, e.g., `units/demo-org/vpc`).
- `--`: Separator. Everything after it is passed directly to the `tofu` or `terraform` binary (e.g., `plan`, `apply`).

To debug, you can prevent the temporary directory from being deleted by using the `-c=false` flag (or by not using `-c`). This allows you to inspect the generated `.tfvars` and the final state of the execution directory.

## Configuration Sources

`iacconsole-cli` supports two sources for infrastructure configuration ("inventory"):

1.  **File-based Inventory**: By default, `iacconsole-cli` reads configuration from JSON files located in the `inventory_path` specified in the `.iacconsolerc` config file. This is the primary method for local development and testing.

2.  **IaCConsole-DB (CMDB)**: If the `IACCONSOLE_API_URL` environment variable is set, `iacconsole-cli` will fetch configuration from this external OpenAPI-based CMDB. This allows for centralized management of configurations.
    -   The `IACCONSOLE_API_URL` contains the API endpoint and credentials.
    -   To get a free account and API keys, visit [https://iacconsole.com/](https://iacconsole.com/), fill in the form with your Account Name, Email, and press Create Account.
    -   You'll receive generated credentials and a ready-to-use export command like `export IACCONSOLE_API_URL=https://6634b72292e9e996105de19e:generatedpassword@api.iacconsole.com`.
    -   Full API documentation is available at [Swagger API docs](https://app.swaggerhub.com/apis-docs/altuhovsu/iacconsole-api/.
    -   The `inventory-to-toaster.sh` script in `examples/` shows how to upload local inventory files to the database.
    -   The CMDB can also be accessed directly from CI/CD pipelines like Jenkins:

```groovy
// Example Jenkins usage to fetch available dimension values
import groovy.json.JsonSlurper

def toasterDimValuesRequest(String dimkey){
    def accessToken = "youraccountid:yourpassword".bytes.encodeBase64().toString()
    def req = new URL("https://api.iacconsole.com/v1/dimension/demo-org/${dimkey}").openConnection();
    req.setRequestProperty("Authorization", "Basic " + accessToken)
    def content = req.getInputStream().getText()
    json = new JsonSlurper().parseText(content)
    return json.Dimensions.DimValue
}

println(toasterDimValuesRequest("datacenter")) // Returns [staging1, staging2]
println(toasterDimValuesRequest("account"))    // Returns [test-account]
```

## Project-Specific Conventions

- **Go Code**: The Go code is straightforward and follows standard practices. The core logic involves file system operations (`os`, `path/filepath`), JSON parsing (`encoding/json`), and command execution (`os/exec`).
- **Variable Injection**: `iacconsole-cli` injects variables into the Terraform context:
  - For a dimension `-d datacenter:staging1`, it creates `var.iacconsole_datacenter_name`, `var.iacconsole_datacenter_data`, and `var.iacconsole_datacenter_defaults`.
  - Environment variables prefixed with `iacconsole-cli_envvar_` (e.g., `export iacconsole-cli_envvar_aws_region=us-east-1`) are available in Terraform as `var.iacconsole_envvar_aws_region`.
- **Configuration Files**:
  - `unit_manifest.json` inside a unit directory can declare required dimensions.
  - `dim_defaults.json` inside an inventory dimension directory provides default values for all items in that dimension.
- **Backend State**: Remote state configuration is managed via the `.iacconsolerc` config file. The state path (`$iacconsole_state_path`) is dynamically generated based on the organization and dimensions to ensure state isolation.
