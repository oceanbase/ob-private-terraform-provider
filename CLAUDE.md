# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Terraform provider project (`ob-terraform-provider`). It is written in Go using the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

## Language Policy

**All user-facing output documents and code comments must be in English**, including but not limited to:

- README, contributing guide, changelogs
- Provider user documentation (`docs/`)
- Comments in example configurations (`examples/`)
- Code comments in Go source files
- Test descriptions and string literals

Technical identifiers (variable names, function names, type names, API paths, etc.) stay in English as usual.

Exception: JSON tags, protobuf fields, external API field names, and anything that must match upstream sources should be kept verbatim.

## Commands

The project uses a Makefile to manage the build workflow:

```sh
# View all available commands
make help

# Build
make build

# Build and install to the local Terraform plugin directory
make install

# Run unit tests
make test

# Run a single test
go test ./internal/provider/... -run TestFunctionName -v

# Run acceptance tests (requires OCP_URL / OCP_USERNAME / OCP_PASSWORD)
make testacc

# Format code & static checks
make fmt
make lint

# Clean build artifacts
make clean
```

## Architecture

Update this section as the provider takes shape. Standard Terraform provider structure:

- `main.go` — entry point; calls `provider.Serve()`
- `internal/provider/` — provider definition, resource and data source implementations
- `examples/` — example Terraform configurations
- `docs/` — generated provider documentation (via `tfplugindocs`)

Each resource typically lives in its own file (`internal/provider/<resource_name>_resource.go`) and implements the `resource.Resource` interface. Data sources follow the same pattern with `_data_source.go`.
