package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSNetworkAclRule_basic(t *testing.T) {
	var networkAcl ec2.NetworkAcl

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNetworkAclRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNetworkAclRuleBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNetworkAclRuleExists("aws_network_acl_rule.baz", &networkAcl),
					testAccCheckAWSNetworkAclRuleExists("aws_network_acl_rule.qux", &networkAcl),
					testAccCheckAWSNetworkAclRuleExists("aws_network_acl_rule.wibble", &networkAcl),
				),
			},
		},
	})
}

func TestAccAWSNetworkAclRule_missingParam(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNetworkAclRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSNetworkAclRuleMissingParam,
				ExpectError: regexp.MustCompile("Either `cidr_block` or `ipv6_cidr_block` must be defined"),
			},
		},
	})
}

func TestAccAWSNetworkAclRule_ipv6(t *testing.T) {
	var networkAcl ec2.NetworkAcl

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNetworkAclRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNetworkAclRuleIpv6Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNetworkAclRuleExists("aws_network_acl_rule.baz", &networkAcl),
				),
			},
		},
	})
}

func TestAccAWSNetworkAclRule_ipv6ICMP(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_network_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNetworkAclRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNetworkAclRuleConfigIpv6ICMP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNetworkAclRuleExists(resourceName, &networkAcl),
				),
			},
		},
	})
}

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/6710
func TestAccAWSNetworkAclRule_ipv6VpcAssignGeneratedIpv6CidrBlockUpdate(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	var vpc ec2.Vpc
	vpcResourceName := "aws_vpc.test"
	resourceName := "aws_network_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNetworkAclRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNetworkAclRuleConfigIpv6VpcAssignGeneratedIpv6CidrBlockUpdate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(vpcResourceName, &vpc),
					resource.TestCheckResourceAttr(vpcResourceName, "assign_generated_ipv6_cidr_block", "false"),
					resource.TestCheckResourceAttr(vpcResourceName, "ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccAWSNetworkAclRuleConfigIpv6VpcAssignGeneratedIpv6CidrBlockUpdate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(vpcResourceName, &vpc),
					resource.TestCheckResourceAttr(vpcResourceName, "assign_generated_ipv6_cidr_block", "true"),
					resource.TestMatchResourceAttr(vpcResourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
					testAccCheckAWSNetworkAclRuleExists(resourceName, &networkAcl),
				),
			},
		},
	})
}

func TestAccAWSNetworkAclRule_allProtocol(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNetworkAclRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccAWSNetworkAclRuleAllProtocolConfig,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:             testAccAWSNetworkAclRuleAllProtocolConfigNoRealUpdate,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccAWSNetworkAclRule_tcpProtocol(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNetworkAclRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccAWSNetworkAclRuleTcpProtocolConfig,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:             testAccAWSNetworkAclRuleTcpProtocolConfigNoRealUpdate,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestResourceAWSNetworkAclRule_validateICMPArgumentValue(t *testing.T) {
	type testCases struct {
		Value    string
		ErrCount int
	}

	invalidCases := []testCases{
		{
			Value:    "",
			ErrCount: 1,
		},
		{
			Value:    "not-a-number",
			ErrCount: 1,
		},
		{
			Value:    "1.0",
			ErrCount: 1,
		},
	}

	for _, tc := range invalidCases {
		_, errors := validateICMPArgumentValue(tc.Value, "icmp_type")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q to trigger a validation error.", tc.Value)
		}
	}

	validCases := []testCases{
		{
			Value:    "0",
			ErrCount: 0,
		},
		{
			Value:    "-1",
			ErrCount: 0,
		},
		{
			Value:    "1",
			ErrCount: 0,
		},
	}

	for _, tc := range validCases {
		_, errors := validateICMPArgumentValue(tc.Value, "icmp_type")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q not to trigger a validation error.", tc.Value)
		}
	}

}

func TestAccAWSNetworkAclRule_deleteRule(t *testing.T) {
	var networkAcl ec2.NetworkAcl

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNetworkAclRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNetworkAclRuleBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNetworkAclRuleExists("aws_network_acl_rule.baz", &networkAcl),
					testAccCheckAWSNetworkAclRuleDelete("aws_network_acl_rule.baz"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSNetworkAclRuleDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		if rs.Type != "aws_network_acl_rule" {
			continue
		}

		req := &ec2.DescribeNetworkAclsInput{
			NetworkAclIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeNetworkAcls(req)
		if err == nil {
			if len(resp.NetworkAcls) > 0 && *resp.NetworkAcls[0].NetworkAclId == rs.Primary.ID {
				networkAcl := resp.NetworkAcls[0]
				if networkAcl.Entries != nil {
					return fmt.Errorf("Network ACL Entries still exist")
				}
			}
		}

		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidNetworkAclID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSNetworkAclRuleExists(n string, networkAcl *ec2.NetworkAcl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network ACL Rule Id is set")
		}

		req := &ec2.DescribeNetworkAclsInput{
			NetworkAclIds: []*string{aws.String(rs.Primary.Attributes["network_acl_id"])},
		}
		resp, err := conn.DescribeNetworkAcls(req)
		if err != nil {
			return err
		}
		if len(resp.NetworkAcls) != 1 {
			return fmt.Errorf("Network ACL not found")
		}
		egress, err := strconv.ParseBool(rs.Primary.Attributes["egress"])
		if err != nil {
			return err
		}
		ruleNo, err := strconv.ParseInt(rs.Primary.Attributes["rule_number"], 10, 64)
		if err != nil {
			return err
		}
		for _, e := range resp.NetworkAcls[0].Entries {
			if *e.RuleNumber == ruleNo && *e.Egress == egress {
				return nil
			}
		}
		return fmt.Errorf("Entry not found: %s", resp.NetworkAcls[0])
	}
}

func testAccCheckAWSNetworkAclRuleDelete(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network ACL Rule Id is set")
		}

		egress, err := strconv.ParseBool(rs.Primary.Attributes["egress"])
		if err != nil {
			return err
		}
		ruleNo, err := strconv.ParseInt(rs.Primary.Attributes["rule_number"], 10, 64)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		_, err = conn.DeleteNetworkAclEntry(&ec2.DeleteNetworkAclEntryInput{
			NetworkAclId: aws.String(rs.Primary.Attributes["network_acl_id"]),
			RuleNumber:   aws.Int64(ruleNo),
			Egress:       aws.Bool(egress),
		})
		if err != nil {
			return fmt.Errorf("Error deleting Network ACL Rule (%s) in testAccCheckAWSNetworkAclRuleDelete: %s", rs.Primary.ID, err)
		}

		return nil
	}
}

const testAccAWSNetworkAclRuleBasicConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.3.0.0/16"
	tags = {
		Name = "terraform-testacc-network-acl-rule-basic"
	}
}

resource "aws_network_acl" "bar" {
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-acl-rule-basic"
	}
}

resource "aws_network_acl_rule" "baz" {
	network_acl_id = "${aws_network_acl.bar.id}"
	rule_number = 200
	egress = false
	protocol = "tcp"
	rule_action = "allow"
	cidr_block = "0.0.0.0/0"
	from_port = 22
	to_port = 22
}

resource "aws_network_acl_rule" "qux" {
	network_acl_id = "${aws_network_acl.bar.id}"
	rule_number = 300
	protocol = "icmp"
	rule_action = "allow"
	cidr_block = "0.0.0.0/0"
	icmp_type = 0
	icmp_code = -1
}

resource "aws_network_acl_rule" "wibble" {
	network_acl_id = "${aws_network_acl.bar.id}"
	rule_number = 400
	protocol = "icmp"
	rule_action = "allow"
	cidr_block = "0.0.0.0/0"
	icmp_type = -1
	icmp_code = -1
}
`

const testAccAWSNetworkAclRuleMissingParam = `
resource "aws_vpc" "foo" {
	cidr_block = "10.3.0.0/16"
	tags = {
		Name = "terraform-testacc-network-acl-rule-missing-param"
	}
}

resource "aws_network_acl" "bar" {
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-acl-rule-missing-param"
	}
}

resource "aws_network_acl_rule" "baz" {
	network_acl_id = "${aws_network_acl.bar.id}"
	rule_number = 200
	egress = false
	protocol = "tcp"
	rule_action = "allow"
	from_port = 22
	to_port = 22
}
`

const testAccAWSNetworkAclRuleAllProtocolConfigNoRealUpdate = `
resource "aws_vpc" "foo" {
	cidr_block = "10.3.0.0/16"
	tags = {
		Name = "terraform-testacc-network-acl-rule-all-proto-no-real-upd"
	}
}

resource "aws_network_acl" "bar" {
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-acl-rule-no-real-update"
	}
}

resource "aws_network_acl_rule" "baz" {
	network_acl_id = "${aws_network_acl.bar.id}"
	rule_number = 150
	egress = false
	protocol = "all"
	rule_action = "allow"
	cidr_block = "0.0.0.0/0"
	from_port = 22
	to_port = 22
}
`

const testAccAWSNetworkAclRuleTcpProtocolConfigNoRealUpdate = `
resource "aws_vpc" "foo" {
	cidr_block = "10.3.0.0/16"
	tags = {
		Name = "testAccAWSNetworkAclRuleTcpProtocolConfigNoRealUpdate"
	}
}
resource "aws_network_acl" "bar" {
	vpc_id = "${aws_vpc.foo.id}"
}
resource "aws_network_acl_rule" "baz" {
	network_acl_id = "${aws_network_acl.bar.id}"
	rule_number = 150
	egress = false
	protocol = "tcp"
	rule_action = "allow"
	cidr_block = "0.0.0.0/0"
	from_port = 22
	to_port = 22
}
`

const testAccAWSNetworkAclRuleAllProtocolConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.3.0.0/16"
	tags = {
		Name = "terraform-testacc-network-acl-rule-proto"
	}
}

resource "aws_network_acl" "bar" {
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-acl-rule-all-protocol"
	}
}

resource "aws_network_acl_rule" "baz" {
	network_acl_id = "${aws_network_acl.bar.id}"
	rule_number = 150
	egress = false
	protocol = "-1"
	rule_action = "allow"
	cidr_block = "0.0.0.0/0"
	from_port = 22
	to_port = 22
}
`

const testAccAWSNetworkAclRuleTcpProtocolConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.3.0.0/16"
	tags = {
		Name = "testAccAWSNetworkAclRuleTcpProtocolConfig"
	}
}
resource "aws_network_acl" "bar" {
	vpc_id = "${aws_vpc.foo.id}"
}
resource "aws_network_acl_rule" "baz" {
	network_acl_id = "${aws_network_acl.bar.id}"
	rule_number = 150
	egress = false
	protocol = "6"
	rule_action = "allow"
	cidr_block = "0.0.0.0/0"
	from_port = 22
	to_port = 22
}
`

const testAccAWSNetworkAclRuleIpv6Config = `
resource "aws_vpc" "foo" {
	cidr_block = "10.3.0.0/16"
	tags = {
		Name = "terraform-testacc-network-acl-rule-ipv6"
	}
}

resource "aws_network_acl" "bar" {
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-acl-rule-ipv6"
	}
}

resource "aws_network_acl_rule" "baz" {
	network_acl_id = "${aws_network_acl.bar.id}"
	rule_number = 150
	egress = false
	protocol = "tcp"
	rule_action = "allow"
	ipv6_cidr_block = "::/0"
	from_port = 22
	to_port = 22
}
`

func testAccAWSNetworkAclRuleConfigIpv6ICMP(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = %q
  }
}

resource "aws_network_acl_rule" "test" {
  from_port       = -1
  icmp_code       = -1
  icmp_type       = -1
  ipv6_cidr_block = "::/0"
  network_acl_id  = "${aws_network_acl.test.id}"
  protocol        = 58
  rule_action     = "allow"
  rule_number     = 150
  to_port         = -1
}
`, rName, rName)
}

func testAccAWSNetworkAclRuleConfigIpv6VpcAssignGeneratedIpv6CidrBlockUpdate(ipv6Enabled bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = %[1]t
  cidr_block                       = "10.3.0.0/16"

  tags = {
    Name = "tf-acc-test-network-acl-rule-ipv6-enabled"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-network-acl-rule-ipv6-enabled"
  }
}

resource "aws_network_acl_rule" "test" {
  count = "${%[1]t ? 1 : 0}"

  from_port       = 22
  ipv6_cidr_block = "${aws_vpc.test.ipv6_cidr_block}"
  network_acl_id  = "${aws_network_acl.test.id}"
  protocol        = "tcp"
  rule_action     = "allow"
  rule_number     = 150
  to_port         = 22
}
`, ipv6Enabled)
}
