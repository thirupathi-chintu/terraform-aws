package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudWatchLogMetricFilter_basic(t *testing.T) {
	var mf cloudwatchlogs.MetricFilter
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogMetricFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogMetricFilterConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogMetricFilterExists("aws_cloudwatch_log_metric_filter.foobar", &mf),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "name", fmt.Sprintf("MyAppAccessCount-%d", rInt)),
					testAccCheckCloudWatchLogMetricFilterName(&mf, fmt.Sprintf("MyAppAccessCount-%d", rInt)),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "pattern", ""),
					testAccCheckCloudWatchLogMetricFilterPattern(&mf, ""),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "log_group_name", fmt.Sprintf("MyApp/access-%d.log", rInt)),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "metric_transformation.0.name", "EventCount"),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "metric_transformation.0.namespace", "YourNamespace"),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "metric_transformation.0.value", "1"),
					testAccCheckCloudWatchLogMetricFilterTransformation(&mf, &cloudwatchlogs.MetricTransformation{
						MetricName:      aws.String("EventCount"),
						MetricNamespace: aws.String("YourNamespace"),
						MetricValue:     aws.String("1"),
					}),
				),
			},
			{
				Config: testAccAWSCloudWatchLogMetricFilterConfigModified(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogMetricFilterExists("aws_cloudwatch_log_metric_filter.foobar", &mf),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "name", fmt.Sprintf("MyAppAccessCount-%d", rInt)),
					testAccCheckCloudWatchLogMetricFilterName(&mf, fmt.Sprintf("MyAppAccessCount-%d", rInt)),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "pattern", "{ $.errorCode = \"AccessDenied\" }"),
					testAccCheckCloudWatchLogMetricFilterPattern(&mf, "{ $.errorCode = \"AccessDenied\" }"),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "log_group_name", fmt.Sprintf("MyApp/access-%d.log", rInt)),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "metric_transformation.0.name", "AccessDeniedCount"),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "metric_transformation.0.namespace", "MyNamespace"),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "metric_transformation.0.value", "2"),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_metric_filter.foobar", "metric_transformation.0.default_value", "1"),
					testAccCheckCloudWatchLogMetricFilterTransformation(&mf, &cloudwatchlogs.MetricTransformation{
						MetricName:      aws.String("AccessDeniedCount"),
						MetricNamespace: aws.String("MyNamespace"),
						MetricValue:     aws.String("2"),
						DefaultValue:    aws.Float64(1),
					}),
				),
			},
			{
				Config: testAccAWSCloudwatchLogMetricFilterConfigMany(rInt),
				Check:  testAccCheckCloudwatchLogMetricFilterManyExist("aws_cloudwatch_log_metric_filter.count_dracula", &mf),
			},
		},
	})
}

func testAccCheckCloudWatchLogMetricFilterName(mf *cloudwatchlogs.MetricFilter, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if name != *mf.FilterName {
			return fmt.Errorf("Expected filter name: %q, given: %q", name, *mf.FilterName)
		}
		return nil
	}
}

func testAccCheckCloudWatchLogMetricFilterPattern(mf *cloudwatchlogs.MetricFilter, pattern string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if mf.FilterPattern == nil {
			if pattern != "" {
				return fmt.Errorf("Received empty filter pattern, expected: %q", pattern)
			}
			return nil
		}

		if pattern != *mf.FilterPattern {
			return fmt.Errorf("Expected filter pattern: %q, given: %q", pattern, *mf.FilterPattern)
		}
		return nil
	}
}

func testAccCheckCloudWatchLogMetricFilterTransformation(mf *cloudwatchlogs.MetricFilter,
	t *cloudwatchlogs.MetricTransformation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		given := mf.MetricTransformations[0]
		expected := t

		if *given.MetricName != *expected.MetricName {
			return fmt.Errorf("Expected metric name: %q, received: %q",
				*expected.MetricName, *given.MetricName)
		}

		if *given.MetricNamespace != *expected.MetricNamespace {
			return fmt.Errorf("Expected metric namespace: %q, received: %q",
				*expected.MetricNamespace, *given.MetricNamespace)
		}

		if *given.MetricValue != *expected.MetricValue {
			return fmt.Errorf("Expected metric value: %q, received: %q",
				*expected.MetricValue, *given.MetricValue)
		}

		if (given.DefaultValue != nil) != (expected.DefaultValue != nil) {
			return fmt.Errorf("Expected default value to be present: %t, received: %t",
				expected.DefaultValue != nil, given.DefaultValue != nil)
		} else if (given.DefaultValue != nil) && *given.DefaultValue != *expected.DefaultValue {
			return fmt.Errorf("Expected metric value: %g, received: %g",
				*expected.DefaultValue, *given.DefaultValue)
		}

		return nil
	}
}

func testAccCheckCloudWatchLogMetricFilterExists(n string, mf *cloudwatchlogs.MetricFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn
		metricFilter, err := lookupCloudWatchLogMetricFilter(conn, rs.Primary.ID, rs.Primary.Attributes["log_group_name"], nil)
		if err != nil {
			return err
		}

		*mf = *metricFilter

		return nil
	}
}

func testAccCheckAWSCloudWatchLogMetricFilterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_metric_filter" {
			continue
		}

		_, err := lookupCloudWatchLogMetricFilter(conn, rs.Primary.ID, rs.Primary.Attributes["log_group_name"], nil)
		if err == nil {
			return fmt.Errorf("MetricFilter Still Exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckCloudwatchLogMetricFilterManyExist(basename string, mf *cloudwatchlogs.MetricFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for i := 0; i < 15; i++ {
			n := fmt.Sprintf("%s.%d", basename, i)
			testfunc := testAccCheckCloudWatchLogMetricFilterExists(n, mf)
			err := testfunc(s)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccAWSCloudWatchLogMetricFilterConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_metric_filter" "foobar" {
  name           = "MyAppAccessCount-%d"
  pattern        = ""
  log_group_name = "${aws_cloudwatch_log_group.dada.name}"

  metric_transformation {
    name      = "EventCount"
    namespace = "YourNamespace"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_group" "dada" {
  name = "MyApp/access-%d.log"
}
`, rInt, rInt)
}

func testAccAWSCloudWatchLogMetricFilterConfigModified(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_metric_filter" "foobar" {
  name = "MyAppAccessCount-%d"

  pattern = <<PATTERN
{ $.errorCode = "AccessDenied" }
PATTERN

  log_group_name = "${aws_cloudwatch_log_group.dada.name}"

  metric_transformation {
    name          = "AccessDeniedCount"
    namespace     = "MyNamespace"
    value         = "2"
    default_value = "1"
  }
}

resource "aws_cloudwatch_log_group" "dada" {
  name = "MyApp/access-%d.log"
}
`, rInt, rInt)
}

func testAccAWSCloudwatchLogMetricFilterConfigMany(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_metric_filter" "count_dracula" {
  count          = 15
  name           = "MyAppCountLog-${count.index}-%d"
  pattern        = "count ${count.index}"
  log_group_name = "${aws_cloudwatch_log_group.mama.name}"

  metric_transformation {
    name      = "CountDracula-${count.index}"
    namespace = "CountNamespace"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_group" "mama" {
  name = "MyApp/count-log-%d.log"
}
`, rInt, rInt)
}
