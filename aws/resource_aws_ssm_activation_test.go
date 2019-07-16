package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSSMActivation_basic(t *testing.T) {
	var ssmActivation ssm.Activation
	name := acctest.RandString(10)
	tag := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMActivationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMActivationBasicConfig(name, tag),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMActivationExists("aws_ssm_activation.foo", &ssmActivation),
					resource.TestCheckResourceAttrSet("aws_ssm_activation.foo", "activation_code"),
					resource.TestCheckResourceAttr("aws_ssm_activation.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_ssm_activation.foo", "tags.Name", tag)),
			},
		},
	})
}

func TestAccAWSSSMActivation_update(t *testing.T) {
	var ssmActivation1, ssmActivation2 ssm.Activation
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMActivationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMActivationBasicConfig(name, "My Activation"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMActivationExists("aws_ssm_activation.foo", &ssmActivation1),
					resource.TestCheckResourceAttrSet("aws_ssm_activation.foo", "activation_code"),
					resource.TestCheckResourceAttr("aws_ssm_activation.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_ssm_activation.foo", "tags.Name", "My Activation"),
				),
			},
			{
				Config: testAccAWSSSMActivationBasicConfig(name, "Foo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMActivationExists("aws_ssm_activation.foo", &ssmActivation2),
					resource.TestCheckResourceAttrSet("aws_ssm_activation.foo", "activation_code"),
					resource.TestCheckResourceAttr("aws_ssm_activation.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_ssm_activation.foo", "tags.Name", "Foo"),
					testAccCheckAWSSSMActivationRecreated(t, &ssmActivation1, &ssmActivation2),
				),
			},
		},
	})
}

func TestAccAWSSSMActivation_expirationDate(t *testing.T) {
	var ssmActivation ssm.Activation
	rName := acctest.RandString(10)
	expirationTime := time.Now().Add(48 * time.Hour)
	expirationDateS := expirationTime.Format(time.RFC3339)
	resourceName := "aws_ssm_activation.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMActivationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSSMActivationConfig_expirationDate(rName, "2018-03-01"),
				ExpectError: regexp.MustCompile(`invalid RFC3339 timestamp`),
			},
			{
				Config: testAccAWSSSMActivationConfig_expirationDate(rName, expirationDateS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMActivationExists(resourceName, &ssmActivation),
					resource.TestCheckResourceAttr(resourceName, "expiration_date", expirationDateS),
				),
			},
		},
	})
}

func testAccCheckAWSSSMActivationRecreated(t *testing.T, before, after *ssm.Activation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.ActivationId == *after.ActivationId {
			t.Fatalf("expected SSM activation Ids to be different but got %v == %v", before.ActivationId, after.ActivationId)
		}
		return nil
	}
}

func testAccCheckAWSSSMActivationExists(n string, ssmActivation *ssm.Activation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Activation ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		resp, err := conn.DescribeActivations(&ssm.DescribeActivationsInput{
			Filters: []*ssm.DescribeActivationsFilter{
				{
					FilterKey: aws.String("ActivationIds"),
					FilterValues: []*string{
						aws.String(rs.Primary.ID),
					},
				},
			},
			MaxResults: aws.Int64(1),
		})

		if err != nil {
			return fmt.Errorf("Could not describe the activation - %s", err)
		}

		*ssmActivation = *resp.ActivationList[0]

		return nil
	}
}

func testAccCheckAWSSSMActivationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_activation" {
			continue
		}

		out, err := conn.DescribeActivations(&ssm.DescribeActivationsInput{
			Filters: []*ssm.DescribeActivationsFilter{
				{
					FilterKey: aws.String("ActivationIds"),
					FilterValues: []*string{
						aws.String(rs.Primary.ID),
					},
				},
			},
			MaxResults: aws.Int64(1),
		})

		if err != nil {
			return err
		}

		if len(out.ActivationList) > 0 {
			return fmt.Errorf("Expected AWS SSM Activation to be gone, but was still found")
		}

		return nil
	}

	return fmt.Errorf("Default error in SSM Activation Test")
}

func testAccAWSSSMActivationBasicConfig(rName string, rTag string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test_role" {
  name = "test_role-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test_attach" {
  role       = "${aws_iam_role.test_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2RoleforSSM"
}

resource "aws_ssm_activation" "foo" {
  name               = "test_ssm_activation-%s"
  description        = "Test"
  iam_role           = "${aws_iam_role.test_role.name}"
  registration_limit = "5"
  depends_on         = ["aws_iam_role_policy_attachment.test_attach"]

  tags = {
    Name = "%s"
  }
}
`, rName, rName, rTag)
}

func testAccAWSSSMActivationConfig_expirationDate(rName, expirationDate string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test_role" {
  name = "test_role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test_attach" {
  role       = "${aws_iam_role.test_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2RoleforSSM"
}

resource "aws_ssm_activation" "foo" {
  name               = "test_ssm_activation-%[1]s"
  description        = "Test"
  expiration_date    = "%[2]s"
  iam_role           = "${aws_iam_role.test_role.name}"
  registration_limit = "5"
  depends_on         = ["aws_iam_role_policy_attachment.test_attach"]
}
`, rName, expirationDate)
}
