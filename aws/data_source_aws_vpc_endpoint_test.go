package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsVpcEndpoint_gatewayBasic(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := fmt.Sprintf("tf-testacc-vpce-%s", acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_gatewayBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttrSet(datasourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttr(datasourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(datasourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
					testAccCheckResourceAttrAccountID(datasourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpoint_byId(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := fmt.Sprintf("tf-testacc-vpce-%s", acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_byId(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttrSet(datasourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttr(datasourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(datasourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
					testAccCheckResourceAttrAccountID(datasourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpoint_gatewayWithRouteTableAndTags(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := fmt.Sprintf("tf-testacc-vpce-%s", acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_gatewayWithRouteTableAndTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttrSet(datasourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "cidr_blocks.#"),
					resource.TestCheckResourceAttr(datasourceName, "route_table_ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(datasourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Name", rName),
					testAccCheckResourceAttrAccountID(datasourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpoint_interface(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint.test"
	rName := fmt.Sprintf("tf-testacc-vpce-%s", acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_interface(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_type", "Interface"),
					resource.TestCheckNoResourceAttr(datasourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(datasourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "network_interface_ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(datasourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Name", rName),
					testAccCheckResourceAttrAccountID(datasourceName, "owner_id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsVpcEndpointConfig_gatewayBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = "${aws_vpc.test.id}"
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
}

data "aws_vpc_endpoint" "test" {
  vpc_id       = "${aws_vpc.test.id}"
  service_name = "${aws_vpc_endpoint.test.service_name}"
  state        = "available"
}
`, rName)
}

func testAccDataSourceAwsVpcEndpointConfig_byId(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = "${aws_vpc.test.id}"
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
}

data "aws_vpc_endpoint" "test" {
  id = "${aws_vpc_endpoint.test.id}"
}
`, rName)
}

func testAccDataSourceAwsVpcEndpointConfig_gatewayWithRouteTableAndTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = "${aws_vpc.test.id}"
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  route_table_ids = [
    "${aws_route_table.test.id}",
  ]

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoint" "test" {
  vpc_id       = "${aws_vpc.test.id}"
  service_name = "${aws_vpc_endpoint.test.service_name}"
  state        = "available"
}
`, rName)
}

func testAccDataSourceAwsVpcEndpointConfig_interface(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_availability_zones" "available" {}

resource "aws_subnet" "test" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "${aws_vpc.test.cidr_block}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = "${aws_vpc.test.id}"
  vpc_endpoint_type   = "Interface"
  service_name        = "com.amazonaws.${data.aws_region.current.name}.ec2"
  private_dns_enabled = false

  subnet_ids = [
    "${aws_subnet.test.id}",
  ]

  security_group_ids = [
    "${aws_security_group.test.id}",
  ]

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_endpoint" "test" {
  vpc_id       = "${aws_vpc.test.id}"
  service_name = "${aws_vpc_endpoint.test.service_name}"
  state        = "available"
}
`, rName)
}
