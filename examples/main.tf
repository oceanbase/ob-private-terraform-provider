# ============================================================================
# OceanBase Terraform Provider example
#
# This example walks through a complete workflow:
#   1. Configure the provider to connect to OCP
#   2. Query available hosts
#   3. Batch-register new hosts to OCP
#   4. Create a 3-replica OB cluster
#   5. Create a business tenant on the cluster
#
# Set environment variables before running:
#   export OCP_URL="http://ocp.example.com:8080"
#   export OCP_USERNAME="admin"
#   export OCP_PASSWORD="your_password"
#
# Then execute:
#   terraform init
#   terraform plan
#   terraform apply
# ============================================================================

terraform {
  required_providers {
    oceanbase = {
      source  = "oceanbase/oceanbase"
      version = "0.1.0"
    }
  }
}

# -------------------------
# Provider configuration
# -------------------------
provider "oceanbase" {
  # Connection parameters are read from OCP_URL / OCP_USERNAME / OCP_PASSWORD env vars.
  # You can also set them inline:
  # ocp_url  = "http://ocp.example.com:8080"
  # username = "admin"
  # password = "your_password"

  task_mode        = "polling" # wait for the async task to finish (recommended)
  polling_timeout  = "30m"
  polling_interval = "10s"
}

# -------------------------
# 1. Query existing AVAILABLE hosts
# -------------------------
data "oceanbase_hosts" "available" {
  status = "AVAILABLE"
}

output "available_hosts" {
  description = "Hosts in OCP whose status is AVAILABLE"
  value       = data.oceanbase_hosts.available.hosts
}

# -------------------------
# 2. Batch-register new hosts to OCP
# -------------------------
# Skip this step if the hosts are already registered in OCP — just use the host IDs
# returned by the data source above.
resource "oceanbase_host" "ob_nodes" {
  hosts = [
    { inner_ip_address = "10.0.0.101" },
    { inner_ip_address = "10.0.0.102" },
    { inner_ip_address = "10.0.0.103" },
  ]

  ssh_port      = 22
  kind          = "DEDICATED_PHYSICAL_MACHINE" # physical machine
  idc_id        = 1                            # IDC ID in OCP
  type_id       = 1                            # host type ID in OCP
  credential_id = 1                            # credential ID in the OCP credential vault
}

output "registered_host_ids" {
  description = "IDs of the newly registered hosts"
  value       = oceanbase_host.ob_nodes.host_ids
}

# -------------------------
# 3. Create a 3-replica OB cluster
# -------------------------
resource "oceanbase_ob_cluster" "production" {
  name = "ob_production"
  type = "PRIMARY"

  # One host per zone, referencing the host IDs registered above.
  zones = [
    {
      name     = "zone1"
      idc_name = "IDCA"
      servers  = [oceanbase_host.ob_nodes.host_ids[0]]
      rpm_name = "oceanbase-4.4.2.1-201000052026040214.el7.x86_64.rpm"
    },
    {
      name     = "zone2"
      idc_name = "IDCA"
      servers  = [oceanbase_host.ob_nodes.host_ids[1]]
      rpm_name = "oceanbase-4.4.2.1-201000052026040214.el7.x86_64.rpm"
    },
    {
      name     = "zone3"
      idc_name = "IDCA"
      servers  = [oceanbase_host.ob_nodes.host_ids[2]]
      rpm_name = "oceanbase-4.4.2.1-201000052026040214.el7.x86_64.rpm"
    },
  ]

  password     = "YourClusterPassword@1"
  primary_zone = "zone1;zone2;zone3"

  # Cluster paths and ports (optional)
  attributes = {
    install_path   = "/home/admin/oceanbase"
    data_disk_path = "/data/1"
    log_disk_path  = "/data/log1"
    sql_port       = 2881
    svr_port       = 2882
  }

  # OB startup parameters (optional)
  startup_parameters = [
    { name = "memory_limit", value = "64G" },
    { name = "system_memory", value = "16G" },
  ]
}

output "cluster_id" {
  description = "ID of the created OB cluster"
  value       = oceanbase_ob_cluster.production.id
}

# -------------------------
# 4. Create a business tenant on the cluster
# -------------------------
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

output "tenant_id" {
  description = "ID of the created tenant"
  value       = oceanbase_ob_tenant.business.id
}
