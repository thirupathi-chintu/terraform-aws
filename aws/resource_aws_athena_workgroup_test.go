package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAthenaWorkGroup_basic(t *testing.T) {
	var workgroup1 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "athena", fmt.Sprintf("workgroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enforce_workgroup_configuration", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.publish_cloudwatch_metrics_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSAthenaWorkGroup_disappears(t *testing.T) {
	var workgroup1 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					testAccCheckAWSAthenaWorkGroupDisappears(&workgroup1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_Configuration_BytesScannedCutoffPerQuery(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigConfigurationBytesScannedCutoffPerQuery(rName, 12582912),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.bytes_scanned_cutoff_per_query", "12582912"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigConfigurationBytesScannedCutoffPerQuery(rName, 10485760),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.bytes_scanned_cutoff_per_query", "10485760"),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_Configuration_EnforceWorkgroupConfiguration(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigConfigurationEnforceWorkgroupConfiguration(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enforce_workgroup_configuration", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigConfigurationEnforceWorkgroupConfiguration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.enforce_workgroup_configuration", "true"),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_Configuration_PublishCloudWatchMetricsEnabled(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigConfigurationPublishCloudWatchMetricsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.publish_cloudwatch_metrics_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigConfigurationPublishCloudWatchMetricsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.publish_cloudwatch_metrics_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_Configuration_ResultConfiguration_EncryptionConfiguration_SseS3(t *testing.T) {
	var workgroup1 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigConfigurationResultConfigurationEncryptionConfigurationEncryptionOptionSseS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.0.encryption_option", athena.EncryptionOptionSseS3),
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

func TestAccAWSAthenaWorkGroup_Configuration_ResultConfiguration_EncryptionConfiguration_Kms(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"
	rEncryption := athena.EncryptionOptionSseKms
	rEncryption2 := athena.EncryptionOptionCseKms

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigConfigurationResultConfigurationEncryptionConfigurationEncryptionOptionWithKms(rName, rEncryption),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.0.encryption_option", rEncryption),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigConfigurationResultConfigurationEncryptionConfigurationEncryptionOptionWithKms(rName, rEncryption2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.encryption_configuration.0.encryption_option", rEncryption2),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_Configuration_ResultConfiguration_OutputLocation(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"
	rOutputLocation1 := fmt.Sprintf("%s-1", rName)
	rOutputLocation2 := fmt.Sprintf("%s-2", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigConfigurationResultConfigurationOutputLocation(rName, rOutputLocation1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation1+"/test/output"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigConfigurationResultConfigurationOutputLocation(rName, rOutputLocation2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.result_configuration.0.output_location", "s3://"+rOutputLocation2+"/test/output"),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_Description(t *testing.T) {
	var workgroup1, workgroup2 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"
	rDescription := acctest.RandString(20)
	rDescriptionUpdate := acctest.RandString(20)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigDescription(rName, rDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "description", rDescription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigDescription(rName, rDescriptionUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "description", rDescriptionUpdate),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_State(t *testing.T) {
	var workgroup1, workgroup2, workgroup3 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigState(rName, athena.WorkGroupStateDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateDisabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigState(rName, athena.WorkGroupStateEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateEnabled),
				),
			},
			{
				Config: testAccAthenaWorkGroupConfigState(rName, athena.WorkGroupStateDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, "state", athena.WorkGroupStateDisabled),
				),
			},
		},
	})
}

func TestAccAWSAthenaWorkGroup_Tags(t *testing.T) {
	var workgroup1, workgroup2, workgroup3 athena.WorkGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_athena_workgroup.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaWorkGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaWorkGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAthenaWorkGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAthenaWorkGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaWorkGroupExists(resourceName, &workgroup3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSAthenaWorkGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).athenaconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_athena_workgroup" {
			continue
		}

		input := &athena.GetWorkGroupInput{
			WorkGroup: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetWorkGroup(input)

		if isAWSErr(err, athena.ErrCodeInvalidRequestException, "is not found") {
			continue
		}

		if err != nil {
			return err
		}

		if resp.WorkGroup != nil {
			return fmt.Errorf("Athena WorkGroup (%s) found", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAWSAthenaWorkGroupExists(name string, workgroup *athena.WorkGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).athenaconn

		input := &athena.GetWorkGroupInput{
			WorkGroup: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetWorkGroup(input)

		if err != nil {
			return err
		}

		*workgroup = *output.WorkGroup

		return nil
	}
}

func testAccCheckAWSAthenaWorkGroupDisappears(workgroup *athena.WorkGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).athenaconn

		input := &athena.DeleteWorkGroupInput{
			WorkGroup: workgroup.Name,
		}

		_, err := conn.DeleteWorkGroup(input)

		return err
	}
}

func testAccAthenaWorkGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAthenaWorkGroupConfigDescription(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  description = %[2]q
  name        = %[1]q
}
`, rName, description)
}

func testAccAthenaWorkGroupConfigConfigurationBytesScannedCutoffPerQuery(rName string, bytesScannedCutoffPerQuery int) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    bytes_scanned_cutoff_per_query = %[2]d
  }
}
`, rName, bytesScannedCutoffPerQuery)
}

func testAccAthenaWorkGroupConfigConfigurationEnforceWorkgroupConfiguration(rName string, enforceWorkgroupConfiguration bool) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    enforce_workgroup_configuration = %[2]t
  }
}
`, rName, enforceWorkgroupConfiguration)
}

func testAccAthenaWorkGroupConfigConfigurationPublishCloudWatchMetricsEnabled(rName string, publishCloudwatchMetricsEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    publish_cloudwatch_metrics_enabled = %[2]t
  }
}
`, rName, publishCloudwatchMetricsEnabled)
}

func testAccAthenaWorkGroupConfigConfigurationResultConfigurationOutputLocation(rName string, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test"{
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    result_configuration {
      output_location = "s3://${aws_s3_bucket.test.bucket}/test/output"
    }
  }
}
`, rName, bucketName)
}

func testAccAthenaWorkGroupConfigConfigurationResultConfigurationEncryptionConfigurationEncryptionOptionSseS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    result_configuration {
      encryption_configuration {
        encryption_option = "SSE_S3"
      }
    }
  }
}
`, rName)
}

func testAccAthenaWorkGroupConfigConfigurationResultConfigurationEncryptionConfigurationEncryptionOptionWithKms(rName, encryptionOption string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_athena_workgroup" "test" {
  name = %[1]q

  configuration {
    result_configuration {
      encryption_configuration {
        encryption_option = %[2]q
        kms_key_arn       = "${aws_kms_key.test.arn}"
      }
    }
  }
}
`, rName, encryptionOption)
}

func testAccAthenaWorkGroupConfigState(rName, state string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name  = %[1]q
  state = %[2]q
}
`, rName, state)
}

func testAccAthenaWorkGroupConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAthenaWorkGroupConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_athena_workgroup" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
