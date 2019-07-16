package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafregional"
)

func TestAccAWSWafRegionalWebAclAssociation_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWafRegionalWebAclAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckWafRegionalWebAclAssociationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWafRegionalWebAclAssociationExists("aws_wafregional_web_acl_association.foo"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalWebAclAssociation_multipleAssociations(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWafRegionalWebAclAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckWafRegionalWebAclAssociationConfig_multipleAssociations,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWafRegionalWebAclAssociationExists("aws_wafregional_web_acl_association.foo"),
					testAccCheckWafRegionalWebAclAssociationExists("aws_wafregional_web_acl_association.bar"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalWebAclAssociation_ResourceArn_ApiGatewayStage(t *testing.T) {
	resourceName := "aws_wafregional_web_acl_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWafRegionalWebAclAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckWafRegionalWebAclAssociationConfigResourceArnApiGatewayStage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWafRegionalWebAclAssociationExists(resourceName),
				),
			},
		},
	})
}

func testAccCheckWafRegionalWebAclAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).wafregionalconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_web_acl_association" {
			continue
		}

		_, resourceArn := resourceAwsWafRegionalWebAclAssociationParseId(rs.Primary.ID)

		input := &wafregional.GetWebACLForResourceInput{
			ResourceArn: aws.String(resourceArn),
		}

		_, err := conn.GetWebACLForResource(input)

		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Resource (%s) still associated to WAF Regional Web ACL", resourceArn)
	}

	return nil
}

func testAccCheckWafRegionalWebAclAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WebACL association ID is set")
		}

		_, resourceArn := resourceAwsWafRegionalWebAclAssociationParseId(rs.Primary.ID)

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn

		input := &wafregional.GetWebACLForResourceInput{
			ResourceArn: aws.String(resourceArn),
		}

		_, err := conn.GetWebACLForResource(input)

		return err
	}
}

const testAccCheckWafRegionalWebAclAssociationConfig_basic = `
resource "aws_wafregional_rule" "foo" {
  name = "foo"
  metric_name = "foo"
}

resource "aws_wafregional_web_acl" "foo" {
  name = "foo"
  metric_name = "foo"
  default_action {
    type = "ALLOW"
  }
  rule {
    action {
      type = "COUNT"
    }
    priority = 100
    rule_id = "${aws_wafregional_rule.foo.id}"
  }
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
}

data "aws_availability_zones" "available" {}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "10.1.1.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
}

resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "10.1.2.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
}

resource "aws_alb" "foo" {
  internal = true
  subnets = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
}

resource "aws_wafregional_web_acl_association" "foo" {
  resource_arn = "${aws_alb.foo.arn}"
  web_acl_id = "${aws_wafregional_web_acl.foo.id}"
}
`

const testAccCheckWafRegionalWebAclAssociationConfig_multipleAssociations = testAccCheckWafRegionalWebAclAssociationConfig_basic + `
resource "aws_alb" "bar" {
  internal = true
  subnets = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
}

resource "aws_wafregional_web_acl_association" "bar" {
  resource_arn = "${aws_alb.bar.arn}"
  web_acl_id = "${aws_wafregional_web_acl.foo.id}"
}
`

func testAccCheckWafRegionalWebAclAssociationConfigResourceArnApiGatewayStage(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  parent_id   = "${aws_api_gateway_rest_api.test.root_resource_id}"
  path_part   = "test"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
}

resource "aws_api_gateway_method" "test" {
  authorization = "NONE"
  http_method   = "GET"
  resource_id   = "${aws_api_gateway_resource.test.id}"
  rest_api_id   = "${aws_api_gateway_rest_api.test.id}"
}

resource "aws_api_gateway_method_response" "test" {
  http_method = "${aws_api_gateway_method.test.http_method}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  http_method             = "${aws_api_gateway_method.test.http_method}"
  integration_http_method = "GET"
  resource_id             = "${aws_api_gateway_resource.test.id}"
  rest_api_id             = "${aws_api_gateway_rest_api.test.id}"
  type                    = "HTTP"
  uri                     = "http://www.example.com"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_integration.test.http_method}"
  status_code = "${aws_api_gateway_method_response.test.status_code}"
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = ["aws_api_gateway_integration_response.test"]

  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
}

resource "aws_api_gateway_stage" "test" {
  deployment_id = "${aws_api_gateway_deployment.test.id}"
  rest_api_id   = "${aws_api_gateway_rest_api.test.id}"
  stage_name    = "test"
}

resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = "test"

  default_action {
    type = "ALLOW"
  }
}

resource "aws_wafregional_web_acl_association" "test" {
  resource_arn = "arn:${data.aws_partition.current.partition}:apigateway:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:/restapis/${aws_api_gateway_rest_api.test.id}/stages/${aws_api_gateway_stage.test.stage_name}"
  web_acl_id   = "${aws_wafregional_web_acl.test.id}"
}
`, rName)
}
