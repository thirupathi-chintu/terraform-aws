package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_vpc_dhcp_options", &resource.Sweeper{
		Name: "aws_vpc_dhcp_options",
		F:    testSweepVpcDhcpOptions,
	})
}

func testSweepVpcDhcpOptions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	input := &ec2.DescribeDhcpOptionsInput{}

	err = conn.DescribeDhcpOptionsPages(input, func(page *ec2.DescribeDhcpOptionsOutput, lastPage bool) bool {
		for _, dhcpOption := range page.DhcpOptions {
			var defaultDomainNameFound, defaultDomainNameServersFound bool

			domainName := region + ".compute.internal"
			if region == "us-east-1" {
				domainName = "ec2.internal"
			}

			// This skips the default dhcp configurations so they don't get deleted
			for _, dhcpConfiguration := range dhcpOption.DhcpConfigurations {
				if aws.StringValue(dhcpConfiguration.Key) == "domain-name" {
					if len(dhcpConfiguration.Values) != 1 || dhcpConfiguration.Values[0] == nil {
						continue
					}

					if aws.StringValue(dhcpConfiguration.Values[0].Value) == domainName {
						defaultDomainNameFound = true
					}
				} else if aws.StringValue(dhcpConfiguration.Key) == "domain-name-servers" {
					if len(dhcpConfiguration.Values) != 1 || dhcpConfiguration.Values[0] == nil {
						continue
					}

					if aws.StringValue(dhcpConfiguration.Values[0].Value) == "AmazonProvidedDNS" {
						defaultDomainNameServersFound = true
					}
				}
			}

			if defaultDomainNameFound && defaultDomainNameServersFound {
				continue
			}

			input := &ec2.DeleteDhcpOptionsInput{
				DhcpOptionsId: dhcpOption.DhcpOptionsId,
			}

			_, err := conn.DeleteDhcpOptions(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 DHCP Option (%s): %s", aws.StringValue(dhcpOption.DhcpOptionsId), err)
			}
		}
		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 DHCP Option sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing DHCP Options: %s", err)
	}

	return nil
}

func TestAccAWSDHCPOptions_importBasic(t *testing.T) {
	resourceName := "aws_vpc_dhcp_options.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDHCPOptions_basic(t *testing.T) {
	var d ec2.DhcpOptions

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists("aws_vpc_dhcp_options.foo", &d),
					resource.TestCheckResourceAttr("aws_vpc_dhcp_options.foo", "domain_name", "service.consul"),
					resource.TestCheckResourceAttr("aws_vpc_dhcp_options.foo", "domain_name_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr("aws_vpc_dhcp_options.foo", "domain_name_servers.1", "10.0.0.2"),
					resource.TestCheckResourceAttr("aws_vpc_dhcp_options.foo", "ntp_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr("aws_vpc_dhcp_options.foo", "netbios_name_servers.0", "127.0.0.1"),
					resource.TestCheckResourceAttr("aws_vpc_dhcp_options.foo", "netbios_node_type", "2"),
					resource.TestCheckResourceAttr("aws_vpc_dhcp_options.foo", "tags.Name", "foo-name"),
					testAccCheckResourceAttrAccountID("aws_vpc_dhcp_options.foo", "owner_id"),
				),
			},
		},
	})
}

func TestAccAWSDHCPOptions_deleteOptions(t *testing.T) {
	var d ec2.DhcpOptions

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDHCPOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPOptionsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists("aws_vpc_dhcp_options.foo", &d),
					testAccCheckDHCPOptionsDelete("aws_vpc_dhcp_options.foo"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDHCPOptionsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_dhcp_options" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeDhcpOptions(&ec2.DescribeDhcpOptionsInput{
			DhcpOptionsIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})
		if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidDhcpOptionID.NotFound" {
			continue
		}
		if err == nil {
			if len(resp.DhcpOptions) > 0 {
				return fmt.Errorf("still exists")
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidDhcpOptionsID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckDHCPOptionsExists(n string, d *ec2.DhcpOptions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeDhcpOptions(&ec2.DescribeDhcpOptionsInput{
			DhcpOptionsIds: []*string{
				aws.String(rs.Primary.ID),
			},
		})
		if err != nil {
			return err
		}
		if len(resp.DhcpOptions) == 0 {
			return fmt.Errorf("DHCP Options not found")
		}

		*d = *resp.DhcpOptions[0]

		return nil
	}
}

func testAccCheckDHCPOptionsDelete(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		_, err := conn.DeleteDhcpOptions(&ec2.DeleteDhcpOptionsInput{
			DhcpOptionsId: aws.String(rs.Primary.ID),
		})

		return err
	}
}

const testAccDHCPOptionsConfig = `
resource "aws_vpc_dhcp_options" "foo" {
	domain_name = "service.consul"
	domain_name_servers = ["127.0.0.1", "10.0.0.2"]
	ntp_servers = ["127.0.0.1"]
	netbios_name_servers = ["127.0.0.1"]
	netbios_node_type = 2

	tags = {
		Name = "foo-name"
	}
}
`
