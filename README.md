# OceanBase Terraform Provider

A Terraform Provider for managing OceanBase On-Premise resources, built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework). It performs full lifecycle management of OceanBase clusters and related components through the OCP (OceanBase Cloud Platform) REST API.

## Table of Contents

- [Requirements](#requirements)
- [Build and Install](#build-and-install)
- [Testing](#testing)
- [Quick Start](#quick-start)
- [Provider Configuration](#provider-configuration)
- [Resources](#resources)
- [Data Sources](#data-sources)
- [Complete Usage Example](#complete-usage-example)

## Requirements

| Dependency | Version |
|------|----------|
| Go | 1.21+ |
| Terraform | 1.x |
| OCP | 4.x (Community Edition or Enterprise Edition) |

## Build and Install

The project uses a Makefile to manage the build process. Run `make help` to see all available commands.

```bash
# View all make commands
make help

# Build the Provider binary
make build

# Build and install to the local Terraform plugin directory (~/.terraform.d/plugins/...)
make install

# Remove the locally installed Provider
make uninstall

# Clean build artifacts
make clean
```

`make install` automatically installs the build output to the `~/.terraform.d/plugins/registry.terraform.io/oceanbase/oceanbase/<version>/<os_arch>/` directory, which Terraform recognizes directly without any additional `~/.terraformrc` configuration.

If you install via `go install .`, you need to configure `dev_overrides` in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "oceanbase/oceanbase" = "/path/to/your/go/bin"
  }
  direct {}
}
```

## Testing

```bash
# Run unit tests
make test

# Run a single test
go test ./internal/provider/... -run TestFunctionName -v

# Run acceptance tests (requires a real OCP environment)
export OCP_URL="http://ocp-host:8080"
export OCP_USERNAME="admin"
export OCP_PASSWORD="your_password"
make testacc
```

## Quick Start

```hcl
terraform {
  required_providers {
    oceanbase = {
      source = "oceanbase/oceanbase"
    }
  }
}

provider "oceanbase" {
  ocp_url  = "http://ocp.example.com:8080"
  username = "admin"
  password = "your_password"
}
```

## Provider Configuration

The Provider communicates with the OCP REST API using OCP's Basic Auth authentication.

```hcl
provider "oceanbase" {
  ocp_url  = "http://ocp.example.com:8080"
  username = "admin"
  password = "your_password"

  # Task mode: polling (poll until task completes) or fire_and_forget (submit only, do not wait)
  task_mode = "polling"

  # Polling timeout (only effective in polling mode)
  polling_timeout = "30m"

  # Polling interval (only effective in polling mode)
  polling_interval = "10s"
}
```

### Configuration Parameters

| Parameter | Type | Required | Description |
|------|------|------|------|
| `ocp_url` | string | Yes | OCP service address, e.g. `http://ocp.example.com:8080` |
| `username` | string | Yes | OCP login username |
| `password` | string | Yes | OCP login password (sensitive) |
| `task_mode` | string | No | Task mode: `polling` (default) or `fire_and_forget` |
| `polling_timeout` | string | No | Polling timeout, default `30m` |
| `polling_interval` | string | No | Polling interval, default `10s` |

### Environment Variables

All connection parameters can be configured via environment variables, which is convenient for CI/CD scenarios:

| Environment Variable | Corresponding Parameter |
|----------|----------|
| `OCP_URL` | `ocp_url` |
| `OCP_USERNAME` | `username` |
| `OCP_PASSWORD` | `password` |

```bash
export OCP_URL="http://ocp.example.com:8080"
export OCP_USERNAME="admin"
export OCP_PASSWORD="your_password"
```

## Resources

The current version (MVP) provides 5 Resources, all of which support Create + Read + Delete operations. Update is not yet supported (i.e. resource changes require destroy and recreate).

---

### oceanbase_ob_cluster

Manages the creation and destruction of an OceanBase cluster. All fields align with the OCP `POST /api/v2/ob/clusters` endpoint.

#### Top-level Parameters

| Parameter | Type | Required | Description |
|------|------|------|------|
| `name` | string | Yes | Cluster name, must match `^[a-zA-Z][a-zA-Z_0-9]{1,47}$` |
| `type` | string | Yes | Cluster type: `PRIMARY` or `STANDBY` |
| `zones` | list(object) | Yes | Zone list, see below |
| `password` | string | No | root@sys password (sensitive) |
| `full_version` | string | No | Full OB version, e.g. `4.2.1.0-100000242023112107` |
| `ob_cluster_id` | number | No | Upstream OB cluster ID |
| `primary_zone` | string | No | Primary Zone, e.g. `zone1;zone2,zone3` |
| `attributes` | object | No | Cluster path and port attributes, see below |
| `startup_parameters` | list(object) | No | OB startup parameters KV list, element `{name, value}` |
| `obproxy_cluster_ids` | list(number) | No | Associated OBProxy cluster IDs |
| `obproxy_user_name` | string | No | OBProxy username (default `proxyro`) |
| `obproxy_user_password` | string | No | OBProxy user password (sensitive) |
| `arbitration_service_id` | number | No | Arbitration service ID |
| `primary_cluster_info` | object | No | Standby cluster info (used when `type=STANDBY`), see below |
| `overselling_factor` | number | No | Resource overselling factor [100, 200] |
| `cgroup_enabled` | bool | No | Whether to enable cgroup (default true) |
| `create_extra_tenant` | bool | No | Whether to create an extra tenant (default false) |
| `support_obs_backup` | bool | No | Whether to support Huawei Cloud OBS backup and restore |
| `load_type` | string | No | Load type |
| `check_id` | string | No | Pre-check ID |
| `client_token` | string | No | Idempotency token, length 8-64 |

**`zones` sub-parameters:**

| Parameter | Type | Required | Description |
|------|------|------|------|
| `name` | string | Yes | Zone name |
| `idc_name` | string | Yes | IDC name |
| `servers` | list(number) | Yes | Host ID list |
| `rpm_name` | string | Yes | Full RPM package filename |
| `architecture` | string | No | `x86_64` or `aarch64` |
| `package_operating_system` | string | No | `el7` or `el8` |
| `root_server` | number | No | rootserver host ID |

**`attributes` sub-parameters:**

| Parameter | Type | Required | Default | Description |
|------|------|------|------|------|
| `install_path` | string | No | `/home/admin/oceanbase` | Installation path |
| `run_path` | string | No | — | Run path |
| `data_disk_path` | string | No | `/data/1` | Data disk path |
| `log_disk_path` | string | No | `/data/log1` | Log disk path |
| `disk_path_style` | string | No | — | Disk path style |
| `operating_system_user` | string | No | — | Host OS user |
| `sql_port` | number | No | 2881 | SQL port |
| `svr_port` | number | No | 2882 | RPC port |

**`primary_cluster_info` sub-parameters:**

| Parameter | Type | Required | Description |
|------|------|------|------|
| `ob_cluster_id` | number | No | Primary cluster OB cluster id |
| `root_sys_password` | string | No | Primary cluster sys password (sensitive) |

**Computed attributes:** `id`, `status`

---

### oceanbase_ob_tenant

Manages the creation and destruction of an OceanBase tenant. All fields align with the OCP `POST /api/v2/ob/clusters/{id}/tenants/createTenant` endpoint.

#### Top-level Parameters

| Parameter | Type | Required | Description |
|------|------|------|------|
| `cluster_id` | number | Yes | Owning cluster ID |
| `name` | string | Yes | Tenant name, must match `^[a-zA-Z][a-zA-Z_0-9]{1,63}$` |
| `root_password` | string | Yes | Tenant root password (sensitive) |
| `zones` | list(object) | Yes | Zone replica configuration, see below |
| `parameters` | list(object) | Yes | Tenant parameter list (may be empty `[]`) |
| `mode` | string | No | `MYSQL` (default) or `ORACLE` |
| `primary_zone` | string | No | Primary Zone (max length 128) |
| `charset` | string | No | Character set |
| `collation` | string | No | Collation |
| `whitelist` | string | No | IP whitelist |
| `time_zone` | string | No | Time zone |
| `description` | string | No | Description (max length 1024) |
| `enable_arbitration` | bool | No | Whether to enable the arbitration service (default false) |
| `skip_import_tenant_info` | bool | No | Whether to skip srs/time_zone import |
| `service_name` | string | No | Tenant service name |
| `load_type` | string | No | Load type |
| `client_token` | string | No | Idempotency token, length 8-64 |

**zones sub-parameters:**

| Parameter | Type | Required | Description |
|------|------|------|------|
| `name` | string | Yes | Zone name |
| `replica_type` | string | Yes | Replica type |
| `resource_pool` | object | Yes | Resource pool configuration |

**resource_pool sub-parameters:**

| Parameter | Type | Required | Description |
|------|------|------|------|
| `unit_spec_name` | string | Yes | Unit spec name |
| `unit_count` | number | Yes | Unit count |

**Computed attributes:** `id`, `status`

---

### oceanbase_obproxy_cluster

Manages the creation and destruction of an OBProxy cluster. All fields align with the OCP `POST /api/v2/obproxy/clusters` endpoint.

#### Top-level Parameters

| Parameter | Type | Required | Description |
|------|------|------|------|
| `name` | string | Yes | OBProxy cluster name |
| `password` | string | No | root@proxysys password (sensitive) |
| `proxyro_password` | string | No | proxyro user password (sensitive) |
| `address` | string | No | Cluster access address |
| `port` | number | No | Cluster access port |
| `work_mode` | string | No | `CONFIG_URL` (default), `RS_LIST`, or `OB_SHARDING` |
| `install_path` | string | No | Installation path |
| `run_path` | string | No | Run path |
| `run_user` | string | No | Run user |
| `obproxy_install_param` | object | No | Installation parameters, see below |
| `ob_links` | list(object) | No | Associated OB cluster list |
| `startup_parameters` | list(object) | No | Startup parameters KV list, element `{name, value}` |
| `parameters` | list(object) | No | Post-startup parameters KV list, element `{name, value}` |
| `client_token` | string | No | Idempotency token, length 8-64 |

**`obproxy_install_param` sub-parameters:**

| Parameter | Type | Required | Description |
|------|------|------|------|
| `host_ids` | list(number) | No | Target host ID list for installation |
| `version` | string | No | Full OBProxy RPM filename |
| `sql_port` | number | No | SQL port |
| `exporter_port` | number | No | Exporter port |
| `rpc_port` | number | No | RPC port (>= 4.3.0) |

**`ob_links` sub-parameters:**

| Parameter | Type | Required | Description |
|------|------|------|------|
| `cluster_name` | string | Yes | OB cluster name |
| `ob_cluster_id` | number | No | OB cluster ID |
| `username` | string | No | Connection user |

**Computed attributes:** `id`, `status`

---

### oceanbase_arbitration_service

Manages the creation and destruction of an arbitration service (Enterprise Edition only). All fields align with the OCP `POST /api/v2/arbitration/services` endpoint.

> **Note:** Calling this resource on Community Edition OCP returns the error: `arbitration_service is only available in OceanBase Enterprise Edition`

#### Parameters

| Parameter | Type | Required | Description |
|------|------|------|------|
| `rpm_name` | string | Yes | RPM package name |
| `host_id` | number | No | Target host ID (>0) |
| `install_path` | string | No | Installation path, default `/home/admin/oceanbase` |
| `run_path` | string | No | Run path, defaults to the same value as `install_path` |
| `svr_port` | number | No | Service port, default 2882 |
| `run_user` | string | No | Run user, default `admin` |
| `description` | string | No | Description |
| `clog_symbolic_link_path` | string | No | clog symbolic link path |
| `startup_parameters` | list(object) | No | Startup parameters KV list, element `{name, value}` |
| `client_token` | string | No | Idempotency token, length 8-64 |

**Computed attributes:** `id`, `status`

---

### oceanbase_host

Batch creates hosts and registers them with OCP. All fields align with the OCP `POST /api/v2/compute/hosts/batchCreate` endpoint. Up to 50 hosts per call.

#### Top-level Parameters

| Parameter | Type | Required | Description |
|------|------|------|------|
| `hosts` | list(object) | Yes | Host list (1-50 entries), see below |
| `ssh_port` | number | Yes | SSH port |
| `kind` | string | Yes | `DEDICATED_PHYSICAL_MACHINE`, `DEDICATED_CONTAINER`, or `DEDICATED_ECS` |
| `publish_ports` | list(string) | No | Used only for `DEDICATED_CONTAINER` |
| `idc_id` | number | Yes | IDC ID |
| `type_id` | number | Yes | Host type ID |
| `credential_id` | number | Yes | Credential ID |
| `alias` | string | No | Alias |
| `description` | string | No | Description |
| `mgragent_port` | number | No | manager agent port |
| `monagent_port` | number | No | monitor agent port |
| `monagent_cpu_quota` | number | No | monitor agent CPU quota (default 1.0) |
| `monagent_memory_quota` | number | No | monitor agent memory (GiB, default 2) |

**`hosts` sub-parameters:**

| Parameter | Type | Required | Description |
|------|------|------|------|
| `inner_ip_address` | string | Yes | Internal IP address |
| `serial_number` | string | No | Serial number |

**Computed attributes:** `id` (synthetic UUID), `host_ids` (list of actually created host IDs)

---

## Data Sources

### oceanbase_hosts

Queries the list of hosts registered in OCP.

#### Parameters

| Parameter | Type | Required | Description |
|------|------|------|------|
| `status` | string | No | Filter by status, e.g. `AVAILABLE` |
| `idc_id` | number | No | Filter by IDC ID |

#### Returned Attributes

The `hosts` list, where each element contains:

| Attribute | Type | Description |
|------|------|------|
| `id` | number | Host ID |
| `inner_ip_address` | string | Internal IP |
| `status` | string | Status |
| `idc_id` | number | IDC ID |
| `type_id` | number | Host type ID |
| `alias` | string | Alias |

---

### oceanbase_ob_clusters

Queries the list of all OB clusters in OCP.

#### Parameters

No filter parameters.

#### Returned Attributes

The `clusters` list, where each element contains:

| Attribute | Type | Description |
|------|------|------|
| `id` | number | Cluster ID |
| `name` | string | Cluster name |
| `status` | string | Status |
| `type` | string | Cluster type |
| `full_version` | string | OB version |

---

## Complete Usage Example

The example below demonstrates a typical OceanBase cluster deployment workflow: starting from querying available hosts, then creating the cluster, tenant, OBProxy cluster, and arbitration service in sequence.

### 1. Provider Configuration

```hcl
terraform {
  required_providers {
    oceanbase = {
      source = "oceanbase/oceanbase"
    }
  }
}

provider "oceanbase" {
  ocp_url  = "http://ocp.example.com:8080"
  username = "admin"
  password = "your_password"

  task_mode        = "polling"
  polling_timeout  = "30m"
  polling_interval = "10s"
}
```

### 2. Query Available Hosts

```hcl
data "oceanbase_hosts" "available" {
  status = "AVAILABLE"
}

output "available_hosts" {
  value = data.oceanbase_hosts.available.hosts
}
```

### 3. Create a Three-Replica OB Cluster

```hcl
resource "oceanbase_ob_cluster" "production" {
  name = "ob_production"
  type = "PRIMARY"

  zones = [
    {
      name     = "zone1"
      idc_name = "idc1"
      servers  = [data.oceanbase_hosts.available.hosts[0].id]
      rpm_name = "oceanbase-4.2.1.0-100000052024010800.el7.x86_64.rpm"
    },
    {
      name     = "zone2"
      idc_name = "idc1"
      servers  = [data.oceanbase_hosts.available.hosts[1].id]
      rpm_name = "oceanbase-4.2.1.0-100000052024010800.el7.x86_64.rpm"
    },
    {
      name     = "zone3"
      idc_name = "idc1"
      servers  = [data.oceanbase_hosts.available.hosts[2].id]
      rpm_name = "oceanbase-4.2.1.0-100000052024010800.el7.x86_64.rpm"
    },
  ]

  password     = "YourClusterPassword@1"
  primary_zone = "zone1;zone2;zone3"
}
```

### 4. Create a Tenant on the Cluster

```hcl
resource "oceanbase_ob_tenant" "business" {
  cluster_id    = oceanbase_ob_cluster.production.id
  name          = "tenant_business"
  root_password = "TenantRootPass@1"
  mode          = "MYSQL"
  primary_zone  = "zone1;zone2;zone3"
  charset       = "utf8mb4"

  zones = [
    {
      name         = "zone1"
      replica_type = "FULL"
      resource_pool = {
        unit_spec_name = "S1"
        unit_count     = 1
      }
    },
    {
      name         = "zone2"
      replica_type = "FULL"
      resource_pool = {
        unit_spec_name = "S1"
        unit_count     = 1
      }
    },
    {
      name         = "zone3"
      replica_type = "FULL"
      resource_pool = {
        unit_spec_name = "S1"
        unit_count     = 1
      }
    },
  ]

  parameters = []
}
```

### 5. Create an OBProxy Cluster

```hcl
resource "oceanbase_obproxy_cluster" "proxy" {
  name = "obproxy_production"

  obproxy_install_param = {
    host_ids      = [data.oceanbase_hosts.available.hosts[3].id]
    version       = "4.2.1.0-1"
    sql_port      = 2883
    exporter_port = 2884
  }

  ob_links = [
    {
      cluster_name  = oceanbase_ob_cluster.production.name
      ob_cluster_id = oceanbase_ob_cluster.production.id
    }
  ]

  password        = "ProxyPass@1"
  proxyro_password = "ProxyroPass@1"
}
```

### 6. Create an Arbitration Service (Enterprise Edition Only)

```hcl
resource "oceanbase_arbitration_service" "arb" {
  rpm_name     = "oceanbase-arbitration-4.2.1.0-100000052024010800.el7.x86_64.rpm"
  host_id      = data.oceanbase_hosts.available.hosts[4].id
  install_path = "/home/admin/arbitration"
  svr_port     = 2882
  run_user     = "admin"
  description  = "Production arbitration service"
}
```

## Notes

- **No Update support**: All Resources in the current MVP version do not support Update operations. When modifying resource attributes, Terraform will perform destroy and recreate.
- **Task mode**: It is recommended to use `polling` mode in production to ensure resources are fully created before proceeding to the next step. The `fire_and_forget` mode can be used for fast validation in CI/CD scenarios.
- **Idempotency token**: All Resources support the `client_token` parameter to guarantee request idempotency. Setting it is recommended in automation scenarios.
- **Enterprise Edition feature**: `oceanbase_arbitration_service` is only available in the OceanBase Enterprise Edition environment.

## License

Please refer to the LICENSE file in the project root directory.
