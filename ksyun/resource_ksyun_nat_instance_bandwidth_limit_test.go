package ksyun

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccKsyunNatInstanceAssociation_basic(t *testing.T) {
	var val map[string]interface{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ksyun_nat_instance_bandwidth_limit.foo",
		Providers:     testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testAccNatInstanceAssociationConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatInstanceBandwidthLimitExists("ksyun_nat_instance_bandwidth_limit.foo", &val),
					testAccCheckNatInstanceBandwidthLimitAttributes(&val),
				),
			},
		},
	})
}

func testAccCheckNatInstanceBandwidthLimitExists(n string, val *map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf(" Nat id is empty ")
		}

		client := testAccProvider.Meta().(*KsyunClient)
		natMap := make(map[string]interface{})
		bwlFilter := BandwidthLimitFilter{
			NetworkInterfaceId: rs.Primary.Attributes["network_interface_id"],
		}
		describeParams := DescribeNatRateLimitParam{
			Filter: bwlFilter,
		}
		err := StructureConverter(describeParams, &natMap)
		if err != nil {
			return err
		}
		natMap["NatId"] = rs.Primary.Attributes["nat_id"]
		ptr, err := client.vpcconn.DescribeNatRateLimit(&natMap)

		if err != nil {
			return err
		}
		if ptr != nil {
			l := (*ptr)["NatNetworkInterfaceSet"].([]interface{})
			if len(l) == 0 {
				return err
			}
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckNatInstanceBandwidthLimitAttributes(val *map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val != nil {
			l := (*val)["NatNetworkInterfaceSet"].([]interface{})
			if len(l) == 0 {
				return fmt.Errorf(" Nat id is empty ")
			}
		}
		return nil
	}
}

// func testAccCheckNatInstanceAssociationDestroy(s *terraform.State) error {
// 	for _, rs := range s.RootModule().Resources {
// 		if rs.Type != "ksyun_nat_instance_bandwidth_limit" {
// 			continue
// 		}
//
// 		client := testAccProvider.Meta().(*KsyunClient)
// 		Nat := make(map[string]interface{})
// 		Nat["NatId.1"] = strings.Split(rs.Primary.ID, ":")[0]
// 		subnetId := strings.Split(rs.Primary.ID, ":")[1]
// 		projectErr := getProjectInfo(&Nat, client)
// 		if projectErr != nil {
// 			return projectErr
// 		}
// 		ptr, err := client.vpcconn.DescribeNats(&Nat)
// 		logger.Debug(logger.ReqFormat, "DescribeNats", ptr)
// 		// Verify the error is what we want
// 		if err != nil {
// 			return err
// 		}
// 		if ptr != nil {
// 			l := (*ptr)["NatSet"].([]interface{})
// 			if len(l) == 1 {
// 				flag := true
// 				if nat, ok := l[0].(map[string]interface{}); ok {
// 					if nat["AssociateInstanceSet"] == nil {
// 						continue
// 					}
// 					if associates, o1 := nat["AssociateInstanceSet"].([]interface{}); o1 {
// 						for _, v := range associates {
// 							if subnet, o2 := v.(map[string]interface{}); o2 {
// 								if subnetId == subnet["NetworkInterfaceId"].(string) {
// 									flag = false
// 									break
// 								}
// 							}
// 						}
// 					}
// 					if flag {
// 						continue
// 					} else {
// 						return fmt.Errorf(" Nat Associate Still Exist ")
// 					}
// 				}
// 			} else {
// 				continue
// 			}
// 		}
// 	}
//
// 	return nil
// }

const testAccNatInstanceAssociationConfig = `

data "ksyun_images" "centos-7_5" {
  platform= "centos-7.5"
}
data "ksyun_availability_zones" "default" {
}

resource "ksyun_security_group" "default" {
  vpc_id = "${ksyun_vpc.foo.id}"
  security_group_name="ksyun-security-group-nat"
}

resource "ksyun_instance" "foo" {
  image_id="${data.ksyun_images.centos-7_5.images.0.image_id}"
  instance_type="N3.2B"

  #max_count=1
  #min_count=1
  subnet_id="${ksyun_subnet.foo.id}"
  instance_password="Xuan663222"
  keep_image_login=false
  charge_type="Daily"
  purchase_time=1
  security_group_id=["${ksyun_security_group.default.id}"]
  instance_name="ksyun-kec-tf-nat"
  sriov_net_support="false"
  project_id=100012
}

resource "ksyun_nat" "foo" {
  nat_name = "ksyun-nat-tf"
  nat_mode = "Subnet"
  nat_type = "public"
  band_width = 200
  charge_type = "DailyPaidByTransfer"
  vpc_id = "${ksyun_vpc.foo.id}"
}
resource "ksyun_vpc" "foo" {
	vpc_name        = "tf-vpc-nat"
	cidr_block = "10.0.5.0/24"
}

resource "ksyun_subnet" "foo" {
  subnet_name      = "tf-acc-nat-subnet1"
  cidr_block = "10.0.5.0/24"
  subnet_type = "Normal"
  vpc_id  = "${ksyun_vpc.foo.id}"
  gateway_ip = "10.0.5.1"
  dns1 = "198.18.254.41"
  dns2 = "198.18.254.40"
  availability_zone = "${data.ksyun_availability_zones.default.availability_zones.0.availability_zone_name}"
}

resource "ksyun_nat_associate" "foo" {
	  nat_id = "${ksyun_nat.foo.id}"
	  network_interface_id = "${ksyun_instance.foo.network_interface_id}"
}

resource "ksyun_nat_instance_bandwidth_limit" "foo" {
  nat_id = "${ksyun_nat.foo.id}"
  network_interface_id = "${ksyun_instance.foo.network_interface_id}"
  bandwidth_limit = 5
}
`
