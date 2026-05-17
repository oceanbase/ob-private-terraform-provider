terraform {
  required_providers {
    oceanbase = {
      source  = "oceanbase/oceanbase"
      version = "0.1.0"
    }
  }
}

provider "oceanbase" {
  # 从环境变量 OCP_URL / OCP_USERNAME / OCP_PASSWORD 读取
}

data "oceanbase_hosts" "available" {
  status = "AVAILABLE"
}

output "available_hosts" {
  value = data.oceanbase_hosts.available.hosts
}
