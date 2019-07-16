package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDBEventSubscription_importBasic(t *testing.T) {
	resourceName := "aws_db_event_subscription.bar"
	rInt := acctest.RandInt()
	subscriptionName := fmt.Sprintf("tf-acc-test-rds-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfig(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     subscriptionName,
			},
		},
	})
}

func TestAccAWSDBEventSubscription_basicUpdate(t *testing.T) {
	var v rds.EventSubscription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-rds-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists("aws_db_event_subscription.bar", &v),
					resource.TestMatchResourceAttr("aws_db_event_subscription.bar", "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:rds:[^:]+:[^:]+:es:%s$", rName))),
					resource.TestCheckResourceAttr("aws_db_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr("aws_db_event_subscription.bar", "source_type", "db-instance"),
					resource.TestCheckResourceAttr("aws_db_event_subscription.bar", "name", rName),
					resource.TestCheckResourceAttr("aws_db_event_subscription.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_db_event_subscription.bar", "tags.Name", "name"),
				),
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists("aws_db_event_subscription.bar", &v),
					resource.TestCheckResourceAttr("aws_db_event_subscription.bar", "enabled", "false"),
					resource.TestCheckResourceAttr("aws_db_event_subscription.bar", "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr("aws_db_event_subscription.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_db_event_subscription.bar", "tags.Name", "new-name"),
				),
			},
		},
	})
}

func TestAccAWSDBEventSubscription_disappears(t *testing.T) {
	var eventSubscription rds.EventSubscription
	rInt := acctest.RandInt()
	resourceName := "aws_db_event_subscription.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &eventSubscription),
					testAccCheckAWSDBEventSubscriptionDisappears(&eventSubscription),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDBEventSubscription_withPrefix(t *testing.T) {
	var v rds.EventSubscription
	rInt := acctest.RandInt()
	startsWithPrefix := regexp.MustCompile("^tf-acc-test-rds-event-subs-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfigWithPrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists("aws_db_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "source_type", "db-instance"),
					resource.TestMatchResourceAttr(
						"aws_db_event_subscription.bar", "name", startsWithPrefix),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "tags.Name", "name"),
				),
			},
		},
	})
}

func TestAccAWSDBEventSubscription_withSourceIds(t *testing.T) {
	var v rds.EventSubscription
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfigWithSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists("aws_db_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "name", fmt.Sprintf("tf-acc-test-rds-event-subs-with-ids-%d", rInt)),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "source_ids.#", "1"),
				),
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigUpdateSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists("aws_db_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "name", fmt.Sprintf("tf-acc-test-rds-event-subs-with-ids-%d", rInt)),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "source_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSDBEventSubscription_categoryUpdate(t *testing.T) {
	var v rds.EventSubscription
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists("aws_db_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "source_type", "db-instance"),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "name", fmt.Sprintf("tf-acc-test-rds-event-subs-%d", rInt)),
				),
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigUpdateCategories(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists("aws_db_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_db_event_subscription.bar", "source_type", "db-instance"),
				),
			},
		},
	})
}

func testAccCheckAWSDBEventSubscriptionExists(n string, v *rds.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS Event Subscription is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		eventSubscription, err := resourceAwsDbEventSubscriptionRetrieve(rs.Primary.ID, conn)

		if err != nil {
			return err
		}

		if eventSubscription == nil {
			return fmt.Errorf("RDS Event Subscription not found")
		}

		*v = *eventSubscription

		return nil
	}
}

func testAccCheckAWSDBEventSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_event_subscription" {
			continue
		}

		eventSubscription, err := resourceAwsDbEventSubscriptionRetrieve(rs.Primary.ID, conn)

		if isAWSErr(err, rds.ErrCodeSubscriptionNotFoundFault, "") {
			continue
		}

		if err != nil {
			return err
		}

		if eventSubscription != nil {
			return fmt.Errorf("RDS Event Subscription (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSDBEventSubscriptionDisappears(eventSubscription *rds.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		input := &rds.DeleteEventSubscriptionInput{
			SubscriptionName: eventSubscription.CustSubscriptionId,
		}

		_, err := conn.DeleteEventSubscription(input)

		if err != nil {
			return err
		}

		return waitForRdsEventSubscriptionDeletion(conn, aws.StringValue(eventSubscription.CustSubscriptionId), 10*time.Minute)
	}
}

func testAccAWSDBEventSubscriptionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%d"
}

resource "aws_db_event_subscription" "bar" {
  name        = "tf-acc-test-rds-event-subs-%d"
  sns_topic   = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type = "db-instance"

  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt, rInt)
}

func testAccAWSDBEventSubscriptionConfigWithPrefix(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%d"
}

resource "aws_db_event_subscription" "bar" {
  name_prefix = "tf-acc-test-rds-event-subs-"
  sns_topic   = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type = "db-instance"

  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt)
}

func testAccAWSDBEventSubscriptionConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%d"
}

resource "aws_db_event_subscription" "bar" {
  name        = "tf-acc-test-rds-event-subs-%d"
  sns_topic   = "${aws_sns_topic.aws_sns_topic.arn}"
  enabled     = false
  source_type = "db-parameter-group"

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "new-name"
  }
}
`, rInt, rInt)
}

func testAccAWSDBEventSubscriptionConfigWithSourceIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%d"
}

resource "aws_db_parameter_group" "bar" {
  name        = "db-parameter-group-event-%d"
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}

resource "aws_db_event_subscription" "bar" {
  name        = "tf-acc-test-rds-event-subs-with-ids-%d"
  sns_topic   = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type = "db-parameter-group"
  source_ids  = ["${aws_db_parameter_group.bar.id}"]

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt, rInt, rInt)
}

func testAccAWSDBEventSubscriptionConfigUpdateSourceIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%d"
}

resource "aws_db_parameter_group" "bar" {
  name        = "db-parameter-group-event-%d"
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}

resource "aws_db_parameter_group" "foo" {
  name        = "db-parameter-group-event-2-%d"
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}

resource "aws_db_event_subscription" "bar" {
  name        = "tf-acc-test-rds-event-subs-with-ids-%d"
  sns_topic   = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type = "db-parameter-group"
  source_ids  = ["${aws_db_parameter_group.bar.id}", "${aws_db_parameter_group.foo.id}"]

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt, rInt, rInt, rInt)
}

func testAccAWSDBEventSubscriptionConfigUpdateCategories(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%d"
}

resource "aws_db_event_subscription" "bar" {
  name        = "tf-acc-test-rds-event-subs-%d"
  sns_topic   = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type = "db-instance"

  event_categories = [
    "availability",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt, rInt)
}
