package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
)

func init() {
	resource.AddTestSweepers("scaleway_lb_private_network", &resource.Sweeper{
		Name:         "scaleway_lb_private_network",
		Dependencies: []string{"scaleway_lb", "scaleway_lb_ip", "scaleway_vpc"},
	})
}

func TestAccScalewayLbPrivateNetwork_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayLbPrivateNetworkDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "test-lb-pn"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_vpc_private_network.pn01", "name"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "test-lb-pn"
					}
			
					resource scaleway_lb_ip ip01 {}
			
					resource scaleway_lb "default" {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb"
						type = "lb-s"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_lb.default", "ip_id"),
					resource.TestCheckResourceAttrSet("scaleway_lb_ip.ip01", "id"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "test-lb-pn"
					}
			
					resource scaleway_lb_ip ip01 {}
			
					resource scaleway_lb "default" {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb"
						type = "lb-s"
					}
			
					resource scaleway_lb_private_network lb01pn01 {
						lb_id = scaleway_lb.default.id
						private_network_id = scaleway_vpc_private_network.pn01.id
						static_config = ["172.16.0.100", "172.16.0.101"]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbPrivateNetworkExists(tt, "scaleway_lb_private_network.lb01pn01"),
					resource.TestCheckResourceAttr("scaleway_lb_private_network.lb01pn01",
						"static_config.0", "172.16.0.100"),
					resource.TestCheckResourceAttr("scaleway_lb_private_network.lb01pn01",
						"static_config.1", "172.16.0.101"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "test-lb-pn"
					}
			
					resource scaleway_lb_ip ip01 {}
			
					resource scaleway_lb "default" {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb2"
						type = "lb-s"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_lb.default", "name"),
					resource.TestCheckResourceAttrSet("scaleway_lb_ip.ip01", "id"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn01 {
						name = "test-lb-without-attachment"
					}
			
					resource scaleway_lb_ip ip01 {}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("scaleway_lb_ip.ip01", "id"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn02 {
						name = "pn_test_network_with_dhcp"
					}

					resource scaleway_vpc_public_gateway_ip gw01 {
					}

					resource scaleway_vpc_public_gateway_dhcp dhcp01 {
						subnet = "192.168.1.0/24"
					}

					resource scaleway_vpc_public_gateway pg01 {
						name = "foobar"
						type = "VPC-GW-S"
						ip_id = scaleway_vpc_public_gateway_ip.gw01.id
					}

					resource scaleway_vpc_gateway_network vpcgw01 {
						gateway_id = scaleway_vpc_public_gateway.pg01.id
						private_network_id = scaleway_vpc_private_network.pn02.id
						dhcp_id = scaleway_vpc_public_gateway_dhcp.dhcp01.id
						cleanup_dhcp = true
						enable_masquerade = true
						depends_on = [scaleway_vpc_public_gateway_ip.gw01, scaleway_vpc_private_network.pn02]
					}
			
					resource scaleway_lb_ip ip02 {}
			
					resource scaleway_lb lb02 {
						ip_id = scaleway_lb_ip.ip02.id
						name = "test-lb"
						type = "lb-s"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_vpc_private_network.pn02", "name", "pn_test_network_with_dhcp"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_ip.gw01", "id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway_dhcp.dhcp01", "subnet"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_public_gateway.pg01", "ip_id"),
					resource.TestCheckResourceAttrSet("scaleway_vpc_gateway_network.vpcgw01", "gateway_id"),
					resource.TestCheckResourceAttrSet("scaleway_lb_ip.ip02", "id"),
					resource.TestCheckResourceAttrSet("scaleway_lb.lb02", "ip_id"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn02 {
						name = "pn_test_network_with_dhcp"
					}
			
					resource scaleway_vpc_public_gateway_ip gw01 {
					}
			
					resource scaleway_vpc_public_gateway_dhcp dhcp01 {
						subnet = "192.168.1.0/24"
					}
			
					resource scaleway_vpc_public_gateway pg01 {
						name = "foobar"
						type = "VPC-GW-S"
						ip_id = scaleway_vpc_public_gateway_ip.gw01.id
					}
			
					resource scaleway_vpc_gateway_network vpcgw01 {
						gateway_id = scaleway_vpc_public_gateway.pg01.id
						private_network_id = scaleway_vpc_private_network.pn02.id
						dhcp_id = scaleway_vpc_public_gateway_dhcp.dhcp01.id
						cleanup_dhcp = true
						enable_masquerade = true
						depends_on = [scaleway_vpc_public_gateway_ip.gw01, scaleway_vpc_private_network.pn02]
					}
			
					resource scaleway_lb_ip ip01 {}
			
					resource scaleway_lb lb02 {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb"
						type = "lb-s"
					}
			
					resource scaleway_lb_private_network lb02pn01 {
						lb_id = scaleway_lb.lb02.id
						private_network_id = scaleway_vpc_private_network.pn02.id
						dhcp_config = true
						depends_on = [scaleway_vpc_public_gateway_dhcp.dhcp01, scaleway_vpc_gateway_network.vpcgw01]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbPrivateNetworkExists(tt, "scaleway_lb_private_network.lb02pn01"),
					resource.TestCheckResourceAttr("scaleway_lb_private_network.lb02pn01", "dhcp_config", "true"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn02 {
						name = "pn_test_network_with_dhcp"
					}
			
					resource scaleway_vpc_public_gateway_ip gw01 {
					}
			
					resource scaleway_vpc_public_gateway_dhcp dhcp01 {
						subnet = "192.168.1.0/24"
					}
			
					resource scaleway_vpc_public_gateway pg01 {
						name = "foobar"
						type = "VPC-GW-S"
						ip_id = scaleway_vpc_public_gateway_ip.gw01.id
					}
			
					resource scaleway_vpc_gateway_network vpcgw01 {
						gateway_id = scaleway_vpc_public_gateway.pg01.id
						private_network_id = scaleway_vpc_private_network.pn02.id
						dhcp_id = scaleway_vpc_public_gateway_dhcp.dhcp01.id
						cleanup_dhcp = true
						enable_masquerade = true
						depends_on = [scaleway_vpc_public_gateway_ip.gw01, scaleway_vpc_private_network.pn02]
					}
			
					resource scaleway_lb_ip ip01 {}
			
					resource scaleway_lb lb02 {
						ip_id = scaleway_lb_ip.ip01.id
						name = "test-lb-dhcp"
						type = "lb-s"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbPrivateNetworkExists(tt, "scaleway_lb_private_network.lb02pn01"),
					resource.TestCheckResourceAttr("scaleway_lb.lb02", "name", "test-lb-dhcp"),
				),
			},
			{
				Config: `
					resource scaleway_vpc_private_network pn02 {
						name = "pn_test_network_with_dhcp"
					}
			
					resource scaleway_vpc_public_gateway_ip gw01 {
					}
			
					resource scaleway_vpc_public_gateway_dhcp dhcp01 {
						subnet = "192.168.1.0/24"
					}
			
					resource scaleway_vpc_public_gateway pg01 {
						name = "foobar"
						type = "VPC-GW-S"
						ip_id = scaleway_vpc_public_gateway_ip.gw01.id
					}
			
					resource scaleway_vpc_gateway_network vpcgw01 {
						gateway_id = scaleway_vpc_public_gateway.pg01.id
						private_network_id = scaleway_vpc_private_network.pn02.id
						dhcp_id = scaleway_vpc_public_gateway_dhcp.dhcp01.id
						cleanup_dhcp = true
						enable_masquerade = true
						depends_on = [scaleway_vpc_public_gateway_ip.gw01, scaleway_vpc_private_network.pn02]
					}

					resource scaleway_lb_ip ip01 {}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayLbPrivateNetworkExists(tt, "scaleway_lb_private_network.lb02pn01"),
					resource.TestCheckResourceAttr("scaleway_vpc_gateway_network.vpcgw01", "cleanup_dhcp", "true"),
				),
			},
		},
	})
}

func getLbPrivateNetwork(tt *TestTools, rs *terraform.ResourceState) (*lb.PrivateNetwork, error) {
	lbID := rs.Primary.Attributes["lb_id"]
	pnID := rs.Primary.Attributes["private_network_id"]

	lbAPI, zone, pnID, err := lbAPIWithZoneAndID(tt.Meta, pnID)
	if err != nil {
		return nil, err
	}

	_, lbID, err = parseZonedID(lbID)
	if err != nil {
		return nil, fmt.Errorf("invalid resource: %s", err)
	}

	listPN, err := lbAPI.ListLBPrivateNetworks(&lb.ZonedAPIListLBPrivateNetworksRequest{
		LBID: lbID,
		Zone: zone,
	})
	if err != nil {
		return nil, err
	}

	for _, pn := range listPN.PrivateNetwork {
		if pn.PrivateNetworkID == pnID {
			return pn, nil
		}
	}

	return nil, fmt.Errorf("private network %s not found", pnID)
}

func testAccCheckScalewayLbPrivateNetworkExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}
		pn, err := getLbPrivateNetwork(tt, rs)
		if err != nil {
			return err
		}
		if pn == nil {
			return fmt.Errorf("resource not found: %s", rs.Primary.Attributes["private_network_id"])
		}

		return nil
	}
}

func testAccCheckScalewayLbPrivateNetworkDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_lb_private_network" {
				continue
			}

			pn, err := getLbPrivateNetwork(tt, rs)
			if err != nil {
				return err
			}

			if pn != nil {
				return fmt.Errorf("LB PN (%s) still exists", rs.Primary.Attributes["private_network_id"])
			}
		}

		return nil
	}
}
