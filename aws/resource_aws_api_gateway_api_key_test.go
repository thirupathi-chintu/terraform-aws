package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGatewayApiKey_basic(t *testing.T) {
	var apiKey1 apigateway.ApiKey
	resourceName := "aws_api_gateway_api_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayApiKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayApiKeyExists(resourceName, &apiKey1),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrSet(resourceName, "value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayApiKey_Description(t *testing.T) {
	var apiKey1, apiKey2 apigateway.ApiKey
	resourceName := "aws_api_gateway_api_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayApiKeyConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayApiKeyExists(resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayApiKeyConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayApiKeyExists(resourceName, &apiKey2),
					testAccCheckAWSAPIGatewayApiKeyNotRecreated(&apiKey1, &apiKey2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayApiKey_Enabled(t *testing.T) {
	var apiKey1, apiKey2 apigateway.ApiKey
	resourceName := "aws_api_gateway_api_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayApiKeyConfigEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayApiKeyExists(resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayApiKeyConfigEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayApiKeyExists(resourceName, &apiKey2),
					testAccCheckAWSAPIGatewayApiKeyNotRecreated(&apiKey1, &apiKey2),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayApiKey_Value(t *testing.T) {
	var apiKey1 apigateway.ApiKey
	resourceName := "aws_api_gateway_api_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayApiKeyConfigValue(rName, `MyCustomToken#@&\"'(§!ç)-_*$€¨^£%ù+=/:.;?,|`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayApiKeyExists(resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, "value", `MyCustomToken#@&\"'(§!ç)-_*$€¨^£%ù+=/:.;?,|`),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGatewayApiKeyExists(n string, res *apigateway.ApiKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway ApiKey ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		req := &apigateway.GetApiKeyInput{
			ApiKey: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetApiKey(req)
		if err != nil {
			return err
		}

		if *describe.Id != rs.Primary.ID {
			return fmt.Errorf("APIGateway ApiKey not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayApiKeyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_api_key" {
			continue
		}

		describe, err := conn.GetApiKeys(&apigateway.GetApiKeysInput{})

		if err == nil {
			if len(describe.Items) != 0 &&
				*describe.Items[0].Id == rs.Primary.ID {
				return fmt.Errorf("API Gateway ApiKey still exists")
			}
		}

		aws2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if aws2err.Code() != "NotFoundException" {
			return err
		}

		return nil
	}

	return nil
}

func testAccCheckAWSAPIGatewayApiKeyNotRecreated(i, j *apigateway.ApiKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreatedDate) != aws.TimeValue(j.CreatedDate) {
			return fmt.Errorf("API Gateway API Key recreated")
		}

		return nil
	}
}

func testAccAWSAPIGatewayApiKeyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAWSAPIGatewayApiKeyConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  description = %[2]q
  name        = %[1]q
}
`, rName, description)
}

func testAccAWSAPIGatewayApiKeyConfigEnabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  enabled = %[2]t
  name    = %[1]q
}
`, rName, enabled)
}

func testAccAWSAPIGatewayApiKeyConfigValue(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  name  = %[1]q
  value = %[2]q
}
`, rName, value)
}
