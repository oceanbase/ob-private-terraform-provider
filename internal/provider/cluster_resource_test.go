package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"oceanbase": providerserver.NewProtocol6WithError(New()()),
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("OCP_URL") == "" || os.Getenv("OCP_USERNAME") == "" || os.Getenv("OCP_PASSWORD") == "" {
		t.Fatal("验收测试需要设置 OCP_URL、OCP_USERNAME、OCP_PASSWORD 环境变量")
	}
}

const testAccClusterConfig = `
resource "oceanbase_ob_cluster" "test" {
  name = "tf_acc_cluster"
  type = "PRIMARY"
  password = "Hello1234!"
  full_version = "4.4.2.1-201000052026040214"
  zones = [
    { name = "zone1", idc_name = "IDCA", servers = [2], rpm_name = "oceanbase-4.4.2.1-201000052026040214.el7.x86_64.rpm", package_operating_system = "el7" },
    { name = "zone2", idc_name = "IDCA", servers = [3], rpm_name = "oceanbase-4.4.2.1-201000052026040214.el7.x86_64.rpm", package_operating_system = "el7" },
    { name = "zone3", idc_name = "IDCA", servers = [8], rpm_name = "oceanbase-4.4.2.1-201000052026040214.el7.x86_64.rpm", package_operating_system = "el7" },
  ]
}
`

func TestAccClusterResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("设置 TF_ACC=1 才会执行验收测试")
	}
	testAccPreCheck(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oceanbase_ob_cluster.test", "name", "tf_acc_cluster"),
					resource.TestCheckResourceAttrSet("oceanbase_ob_cluster.test", "id"),
				),
			},
		},
	})
}
