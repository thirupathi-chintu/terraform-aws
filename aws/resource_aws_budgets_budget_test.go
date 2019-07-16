package aws

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSBudgetsBudget_basic(t *testing.T) {
	costFilterKey := "AZ"
	name := fmt.Sprintf("test-budget-%d", acctest.RandInt())
	configBasicDefaults := testAccAWSBudgetsBudgetConfigDefaults(name)
	accountID := "012345678910"
	configBasicUpdate := testAccAWSBudgetsBudgetConfigUpdate(name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBudgets(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfig_BasicDefaults(configBasicDefaults, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name", regexp.MustCompile(*configBasicDefaults.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", *configBasicDefaults.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", *configBasicDefaults.BudgetLimit.Amount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", *configBasicDefaults.BudgetLimit.Unit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicDefaults.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicDefaults.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", *configBasicDefaults.TimeUnit),
				),
			},
			{
				PlanOnly:    true,
				Config:      testAccAWSBudgetsBudgetConfig_WithAccountID(configBasicDefaults, accountID, costFilterKey),
				ExpectError: regexp.MustCompile("account_id.*" + accountID),
			},
			{
				Config: testAccAWSBudgetsBudgetConfig_Basic(configBasicUpdate, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicUpdate),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name", regexp.MustCompile(*configBasicUpdate.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", *configBasicUpdate.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", *configBasicUpdate.BudgetLimit.Amount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", *configBasicUpdate.BudgetLimit.Unit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicUpdate.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicUpdate.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", *configBasicUpdate.TimeUnit),
				),
			},
			{
				ResourceName:            "aws_budgets_budget.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSBudgetsBudget_prefix(t *testing.T) {
	costFilterKey := "AZ"
	name := "test-budget-"
	configBasicDefaults := testAccAWSBudgetsBudgetConfigDefaults(name)
	configBasicUpdate := testAccAWSBudgetsBudgetConfigUpdate(name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBudgets(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBudgetsBudgetConfig_PrefixDefaults(configBasicDefaults, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name_prefix", regexp.MustCompile(*configBasicDefaults.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", *configBasicDefaults.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", *configBasicDefaults.BudgetLimit.Amount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", *configBasicDefaults.BudgetLimit.Unit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicDefaults.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicDefaults.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", *configBasicDefaults.TimeUnit),
				),
			},

			{
				Config: testAccAWSBudgetsBudgetConfig_Prefix(configBasicUpdate, costFilterKey),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicUpdate),
					resource.TestMatchResourceAttr("aws_budgets_budget.foo", "name_prefix", regexp.MustCompile(*configBasicUpdate.BudgetName)),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "budget_type", *configBasicUpdate.BudgetType),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_amount", *configBasicUpdate.BudgetLimit.Amount),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "limit_unit", *configBasicUpdate.BudgetLimit.Unit),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_start", configBasicUpdate.TimePeriod.Start.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_period_end", configBasicUpdate.TimePeriod.End.Format("2006-01-02_15:04")),
					resource.TestCheckResourceAttr("aws_budgets_budget.foo", "time_unit", *configBasicUpdate.TimeUnit),
				),
			},

			{
				ResourceName:            "aws_budgets_budget.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSBudgetsBudget_notification(t *testing.T) {
	name := fmt.Sprintf("test-budget-%d", acctest.RandInt())
	configBasicDefaults := testAccAWSBudgetsBudgetConfigDefaults(name)
	configBasicDefaults.CostFilters = map[string][]*string{}

	notificationConfigDefaults := []budgets.Notification{testAccAWSBudgetsBudgetNotificationConfigDefaults()}
	notificationConfigUpdated := []budgets.Notification{testAccAWSBudgetsBudgetNotificationConfigUpdate()}
	twoNotificationConfigs := []budgets.Notification{
		testAccAWSBudgetsBudgetNotificationConfigUpdate(),
		testAccAWSBudgetsBudgetNotificationConfigDefaults(),
	}

	noEmails := []string{}
	oneEmail := []string{"foo@example.com"}
	oneOtherEmail := []string{"bar@example.com"}
	twoEmails := []string{"bar@example.com", "baz@example.com"}
	noTopics := []string{}
	oneTopic := []string{"${aws_sns_topic.budget_notifications.arn}"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBudgets(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSBudgetsBudgetDestroy,
		Steps: []resource.TestStep{
			// Can't create without at least one subscriber
			{
				Config:      testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, noEmails, noTopics),
				ExpectError: regexp.MustCompile(`Notification must have at least one subscriber`),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
				),
			},
			// Basic Notification with only email
			{
				Config: testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, oneEmail, noTopics),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
				),
			},
			// Change only subscriber to a different e-mail
			{
				Config: testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, oneOtherEmail, noTopics),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
				),
			},
			// Add a second e-mail and a topic
			{
				Config: testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, twoEmails, oneTopic),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
				),
			},
			// Delete both E-Mails
			{
				Config: testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, noEmails, oneTopic),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
				),
			},
			// Swap one Topic fo one E-Mail
			{
				Config: testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, oneEmail, noTopics),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
				),
			},
			// Can't update without at least one subscriber
			{
				Config:      testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigDefaults, noEmails, noTopics),
				ExpectError: regexp.MustCompile(`Notification must have at least one subscriber`),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
				),
			},
			// Update all non-subscription parameters
			{
				Config:      testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, notificationConfigUpdated, noEmails, noTopics),
				ExpectError: regexp.MustCompile(`Notification must have at least one subscriber`),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
				),
			},
			// Add a second subscription
			{
				Config:      testAccAWSBudgetsBudgetConfigWithNotification_Basic(configBasicDefaults, twoNotificationConfigs, noEmails, noTopics),
				ExpectError: regexp.MustCompile(`Notification must have at least one subscriber`),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSBudgetsBudgetExists("aws_budgets_budget.foo", configBasicDefaults),
				),
			},
		},
	})
}

func testAccAWSBudgetsBudgetExists(resourceName string, config budgets.Budget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		accountID, budgetName, err := decodeBudgetsBudgetID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed decoding ID: %v", err)
		}

		client := testAccProvider.Meta().(*AWSClient).budgetconn
		b, err := client.DescribeBudget(&budgets.DescribeBudgetInput{
			BudgetName: &budgetName,
			AccountId:  &accountID,
		})

		if err != nil {
			return fmt.Errorf("Describebudget error: %v", err)
		}

		if b.Budget == nil {
			return fmt.Errorf("No budget returned %v in %v", b.Budget, b)
		}

		if *b.Budget.BudgetLimit.Amount != *config.BudgetLimit.Amount {
			return fmt.Errorf("budget limit incorrectly set %v != %v", *config.BudgetLimit.Amount, *b.Budget.BudgetLimit.Amount)
		}

		if err := testAccAWSBudgetsBudgetCheckCostTypes(config, *b.Budget.CostTypes); err != nil {
			return err
		}

		if err := testAccAWSBudgetsBudgetCheckTimePeriod(*config.TimePeriod, *b.Budget.TimePeriod); err != nil {
			return err
		}

		if !reflect.DeepEqual(b.Budget.CostFilters, config.CostFilters) {
			return fmt.Errorf("cost filter not set properly: %v != %v", b.Budget.CostFilters, config.CostFilters)
		}

		return nil
	}
}

func testAccAWSBudgetsBudgetCheckTimePeriod(configTimePeriod, timePeriod budgets.TimePeriod) error {
	if configTimePeriod.End.Format("2006-01-02_15:04") != timePeriod.End.Format("2006-01-02_15:04") {
		return fmt.Errorf("TimePeriodEnd not set properly '%v' should be '%v'", *timePeriod.End, *configTimePeriod.End)
	}

	if configTimePeriod.Start.Format("2006-01-02_15:04") != timePeriod.Start.Format("2006-01-02_15:04") {
		return fmt.Errorf("TimePeriodStart not set properly '%v' should be '%v'", *timePeriod.Start, *configTimePeriod.Start)
	}

	return nil
}

func testAccAWSBudgetsBudgetCheckCostTypes(config budgets.Budget, costTypes budgets.CostTypes) error {
	if *costTypes.IncludeCredit != *config.CostTypes.IncludeCredit {
		return fmt.Errorf("IncludeCredit not set properly '%v' should be '%v'", *costTypes.IncludeCredit, *config.CostTypes.IncludeCredit)
	}

	if *costTypes.IncludeOtherSubscription != *config.CostTypes.IncludeOtherSubscription {
		return fmt.Errorf("IncludeOtherSubscription not set properly '%v' should be '%v'", *costTypes.IncludeOtherSubscription, *config.CostTypes.IncludeOtherSubscription)
	}

	if *costTypes.IncludeRecurring != *config.CostTypes.IncludeRecurring {
		return fmt.Errorf("IncludeRecurring not set properly  '%v' should be '%v'", *costTypes.IncludeRecurring, *config.CostTypes.IncludeRecurring)
	}

	if *costTypes.IncludeRefund != *config.CostTypes.IncludeRefund {
		return fmt.Errorf("IncludeRefund not set properly '%v' should be '%v'", *costTypes.IncludeRefund, *config.CostTypes.IncludeRefund)
	}

	if *costTypes.IncludeSubscription != *config.CostTypes.IncludeSubscription {
		return fmt.Errorf("IncludeSubscription not set properly '%v' should be '%v'", *costTypes.IncludeSubscription, *config.CostTypes.IncludeSubscription)
	}

	if *costTypes.IncludeSupport != *config.CostTypes.IncludeSupport {
		return fmt.Errorf("IncludeSupport not set properly '%v' should be '%v'", *costTypes.IncludeSupport, *config.CostTypes.IncludeSupport)
	}

	if *costTypes.IncludeTax != *config.CostTypes.IncludeTax {
		return fmt.Errorf("IncludeTax not set properly '%v' should be '%v'", *costTypes.IncludeTax, *config.CostTypes.IncludeTax)
	}

	if *costTypes.IncludeUpfront != *config.CostTypes.IncludeUpfront {
		return fmt.Errorf("IncludeUpfront not set properly '%v' should be '%v'", *costTypes.IncludeUpfront, *config.CostTypes.IncludeUpfront)
	}

	if *costTypes.UseBlended != *config.CostTypes.UseBlended {
		return fmt.Errorf("UseBlended not set properly '%v' should be '%v'", *costTypes.UseBlended, *config.CostTypes.UseBlended)
	}

	return nil
}

func testAccAWSBudgetsBudgetDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	client := meta.(*AWSClient).budgetconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_budgets_budget" {
			continue
		}

		accountID, budgetName, err := decodeBudgetsBudgetID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Budget '%s': id could not be decoded and could not be deleted properly", rs.Primary.ID)
		}

		_, err = client.DescribeBudget(&budgets.DescribeBudgetInput{
			BudgetName: aws.String(budgetName),
			AccountId:  aws.String(accountID),
		})
		if !isAWSErr(err, budgets.ErrCodeNotFoundException, "") {
			return fmt.Errorf("Budget '%s' was not deleted properly", rs.Primary.ID)
		}
	}

	return nil
}

func testAccPreCheckAWSBudgets(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).budgetconn

	input := &budgets.DescribeBudgetsInput{
		AccountId: aws.String(testAccProvider.Meta().(*AWSClient).accountid),
	}

	_, err := conn.DescribeBudgets(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSBudgetsBudgetConfigUpdate(name string) budgets.Budget {
	dateNow := time.Now().UTC()
	futureDate := dateNow.AddDate(0, 0, 14)
	startDate := dateNow.AddDate(0, 0, -14)
	return budgets.Budget{
		BudgetName: aws.String(name),
		BudgetType: aws.String("COST"),
		BudgetLimit: &budgets.Spend{
			Amount: aws.String("500.0"),
			Unit:   aws.String("USD"),
		},
		CostFilters: map[string][]*string{
			"AZ": {
				aws.String("us-east-2"),
			},
		},
		CostTypes: &budgets.CostTypes{
			IncludeCredit:            aws.Bool(true),
			IncludeOtherSubscription: aws.Bool(true),
			IncludeRecurring:         aws.Bool(true),
			IncludeRefund:            aws.Bool(true),
			IncludeSubscription:      aws.Bool(false),
			IncludeSupport:           aws.Bool(true),
			IncludeTax:               aws.Bool(false),
			IncludeUpfront:           aws.Bool(true),
			UseBlended:               aws.Bool(false),
		},
		TimeUnit: aws.String("MONTHLY"),
		TimePeriod: &budgets.TimePeriod{
			End:   &futureDate,
			Start: &startDate,
		},
	}
}

func testAccAWSBudgetsBudgetConfigDefaults(name string) budgets.Budget {
	dateNow := time.Now().UTC()
	futureDate := time.Date(2087, 6, 15, 00, 0, 0, 0, time.UTC)
	startDate := dateNow.AddDate(0, 0, -14)
	return budgets.Budget{
		BudgetName: aws.String(name),
		BudgetType: aws.String("COST"),
		BudgetLimit: &budgets.Spend{
			Amount: aws.String("100.0"),
			Unit:   aws.String("USD"),
		},
		CostFilters: map[string][]*string{
			"AZ": {
				aws.String("us-east-1"),
			},
		},
		CostTypes: &budgets.CostTypes{
			IncludeCredit:            aws.Bool(true),
			IncludeOtherSubscription: aws.Bool(true),
			IncludeRecurring:         aws.Bool(true),
			IncludeRefund:            aws.Bool(true),
			IncludeSubscription:      aws.Bool(true),
			IncludeSupport:           aws.Bool(true),
			IncludeTax:               aws.Bool(true),
			IncludeUpfront:           aws.Bool(true),
			UseBlended:               aws.Bool(false),
		},
		TimeUnit: aws.String("MONTHLY"),
		TimePeriod: &budgets.TimePeriod{
			End:   &futureDate,
			Start: &startDate,
		},
	}
}

func testAccAWSBudgetsBudgetNotificationConfigDefaults() budgets.Notification {
	return budgets.Notification{
		NotificationType:   aws.String(budgets.NotificationTypeActual),
		ThresholdType:      aws.String(budgets.ThresholdTypeAbsoluteValue),
		Threshold:          aws.Float64(100.0),
		ComparisonOperator: aws.String(budgets.ComparisonOperatorGreaterThan),
	}
}
func testAccAWSBudgetsBudgetNotificationConfigUpdate() budgets.Notification {
	return budgets.Notification{
		NotificationType:   aws.String(budgets.NotificationTypeForecasted),
		ThresholdType:      aws.String(budgets.ThresholdTypePercentage),
		Threshold:          aws.Float64(200.0),
		ComparisonOperator: aws.String(budgets.ComparisonOperatorLessThan),
	}
}

func testAccAWSBudgetsBudgetConfig_WithAccountID(budgetConfig budgets.Budget, accountID, costFilterKey string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	costFilterValue := *budgetConfig.CostFilters[costFilterKey][0]

	return fmt.Sprintf(`
resource "aws_budgets_budget" "foo" {
  account_id        = "%s"
  name_prefix       = "%s"
  budget_type       = "%s"
  limit_amount      = "%s"
  limit_unit        = "%s"
  time_period_start = "%s"
  time_unit         = "%s"

  cost_filters = {
    "%s" = "%s"
  }
}
`, accountID, *budgetConfig.BudgetName, *budgetConfig.BudgetType, *budgetConfig.BudgetLimit.Amount, *budgetConfig.BudgetLimit.Unit, timePeriodStart, *budgetConfig.TimeUnit, costFilterKey, costFilterValue)
}

func testAccAWSBudgetsBudgetConfig_PrefixDefaults(budgetConfig budgets.Budget, costFilterKey string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	costFilterValue := *budgetConfig.CostFilters[costFilterKey][0]

	return fmt.Sprintf(`
resource "aws_budgets_budget" "foo" {
  name_prefix       = "%s"
  budget_type       = "%s"
  limit_amount      = "%s"
  limit_unit        = "%s"
  time_period_start = "%s"
  time_unit         = "%s"

  cost_filters = {
    "%s" = "%s"
  }
}
`, *budgetConfig.BudgetName, *budgetConfig.BudgetType, *budgetConfig.BudgetLimit.Amount, *budgetConfig.BudgetLimit.Unit, timePeriodStart, *budgetConfig.TimeUnit, costFilterKey, costFilterValue)
}

func testAccAWSBudgetsBudgetConfig_Prefix(budgetConfig budgets.Budget, costFilterKey string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	timePeriodEnd := budgetConfig.TimePeriod.End.Format("2006-01-02_15:04")
	costFilterValue := *budgetConfig.CostFilters[costFilterKey][0]

	return fmt.Sprintf(`
resource "aws_budgets_budget" "foo" {
  name_prefix  = "%s"
  budget_type  = "%s"
  limit_amount = "%s"
  limit_unit   = "%s"

  cost_types {
    include_tax          = "%t"
    include_subscription = "%t"
    use_blended          = "%t"
  }

  time_period_start = "%s"
  time_period_end   = "%s"
  time_unit         = "%s"

  cost_filters = {
    "%s" = "%s"
  }
}
`, *budgetConfig.BudgetName, *budgetConfig.BudgetType, *budgetConfig.BudgetLimit.Amount, *budgetConfig.BudgetLimit.Unit, *budgetConfig.CostTypes.IncludeTax, *budgetConfig.CostTypes.IncludeSubscription, *budgetConfig.CostTypes.UseBlended, timePeriodStart, timePeriodEnd, *budgetConfig.TimeUnit, costFilterKey, costFilterValue)
}
func testAccAWSBudgetsBudgetConfig_BasicDefaults(budgetConfig budgets.Budget, costFilterKey string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	costFilterValue := *budgetConfig.CostFilters[costFilterKey][0]

	return fmt.Sprintf(`
resource "aws_budgets_budget" "foo" {
  name              = "%s"
  budget_type       = "%s"
  limit_amount      = "%s"
  limit_unit        = "%s"
  time_period_start = "%s"
  time_unit         = "%s"

  cost_filters = {
    "%s" = "%s"
  }
}
`, *budgetConfig.BudgetName, *budgetConfig.BudgetType, *budgetConfig.BudgetLimit.Amount, *budgetConfig.BudgetLimit.Unit, timePeriodStart, *budgetConfig.TimeUnit, costFilterKey, costFilterValue)
}

func testAccAWSBudgetsBudgetConfig_Basic(budgetConfig budgets.Budget, costFilterKey string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	timePeriodEnd := budgetConfig.TimePeriod.End.Format("2006-01-02_15:04")
	costFilterValue := *budgetConfig.CostFilters[costFilterKey][0]

	return fmt.Sprintf(`
resource "aws_budgets_budget" "foo" {
  name         = "%s"
  budget_type  = "%s"
  limit_amount = "%s"
  limit_unit   = "%s"

  cost_types {
    include_tax          = "%t"
    include_subscription = "%t"
    use_blended          = "%t"
  }

  time_period_start = "%s"
  time_period_end   = "%s"
  time_unit         = "%s"

  cost_filters = {
    "%s" = "%s"
  }
}
`, *budgetConfig.BudgetName, *budgetConfig.BudgetType, *budgetConfig.BudgetLimit.Amount, *budgetConfig.BudgetLimit.Unit, *budgetConfig.CostTypes.IncludeTax, *budgetConfig.CostTypes.IncludeSubscription, *budgetConfig.CostTypes.UseBlended, timePeriodStart, timePeriodEnd, *budgetConfig.TimeUnit, costFilterKey, costFilterValue)
}

func testAccAWSBudgetsBudgetConfigWithNotification_Basic(budgetConfig budgets.Budget, notifications []budgets.Notification, emails []string, topics []string) string {
	timePeriodStart := budgetConfig.TimePeriod.Start.Format("2006-01-02_15:04")
	timePeriodEnd := budgetConfig.TimePeriod.End.Format("2006-01-02_15:04")
	notificationStrings := make([]string, len(notifications))

	for i, notification := range notifications {
		notificationStrings[i] = testAccAWSBudgetsBudgetConfigNotificationSnippet(notification, emails, topics)
	}

	return fmt.Sprintf(`
resource "aws_sns_topic" "budget_notifications" {
  name_prefix = "user-updates-topic"
}

resource "aws_budgets_budget" "foo" {
	name = "%s"
	budget_type = "%s"
	limit_amount = "%s"
	limit_unit = "%s"
	cost_types {
		include_tax = "%t"
		include_subscription = "%t"
		use_blended = "%t"
	}

	time_period_start = "%s" 
	time_period_end = "%s"
	time_unit = "%s"
	
    %s
}
`, *budgetConfig.BudgetName, *budgetConfig.BudgetType, *budgetConfig.BudgetLimit.Amount, *budgetConfig.BudgetLimit.Unit, *budgetConfig.CostTypes.IncludeTax, *budgetConfig.CostTypes.IncludeSubscription, *budgetConfig.CostTypes.UseBlended, timePeriodStart, timePeriodEnd, *budgetConfig.TimeUnit, strings.Join(notificationStrings, "\n"))

}

func testAccAWSBudgetsBudgetConfigNotificationSnippet(notification budgets.Notification, emails []string, topics []string) string {
	quotedEMails := make([]string, len(emails))
	for i, email := range emails {
		quotedEMails[i] = strconv.Quote(email)
	}

	quotedTopics := make([]string, len(topics))
	for i, topic := range topics {
		quotedTopics[i] = strconv.Quote(topic)
	}

	return fmt.Sprintf(`
	notification {
		threshold = %f
		threshold_type = "%s"
		notification_type = "%s"
		subscriber_email_addresses = [%s]
		subscriber_sns_topic_arns = [%s]
		comparison_operator = "%s"
	}
`, *notification.Threshold, *notification.ThresholdType, *notification.NotificationType, strings.Join(quotedEMails, ","), strings.Join(quotedTopics, ","), *notification.ComparisonOperator)
}
