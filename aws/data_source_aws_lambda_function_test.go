package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAWSLambdaFunction_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dead_letter_config.#", resourceName, "dead_letter_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "function_name", resourceName, "function_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "handler", resourceName, "handler"),
					resource.TestCheckResourceAttrPair(dataSourceName, "invoke_arn", resourceName, "invoke_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified", resourceName, "last_modified"),
					resource.TestCheckResourceAttrPair(dataSourceName, "memory_size", resourceName, "memory_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "qualified_arn", resourceName, "qualified_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "reserved_concurrent_executions", resourceName, "reserved_concurrent_executions"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role", resourceName, "role"),
					resource.TestCheckResourceAttrPair(dataSourceName, "runtime", resourceName, "runtime"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_code_hash", resourceName, "source_code_hash"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_code_size", resourceName, "source_code_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "timeout", resourceName, "timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tracing_config.#", resourceName, "tracing_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tracing_config.0.mode", resourceName, "tracing_config.0.mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaFunction_version(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "qualified_arn", resourceName, "qualified_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "qualifier", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "version", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaFunction_alias(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_lambda_function.test"
	lambdaAliasResourceName := "aws_lambda_alias.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigAlias(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "qualified_arn", lambdaAliasResourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "qualifier", lambdaAliasResourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", lambdaAliasResourceName, "function_version"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaFunction_layers(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigLayers(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "layers.#", resourceName, "layers.#"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaFunction_vpc(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_config.#", resourceName, "vpc_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_config.0.security_group_ids.#", resourceName, "vpc_config.0.security_group_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_config.0.subnet_ids.#", resourceName, "vpc_config.0.subnet_ids.#"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaFunction_environment(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigEnvironment(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "environment.#", resourceName, "environment.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "environment.0.variables.%", resourceName, "environment.0.variables.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "environment.0.variables.key1", resourceName, "environment.0.variables.key1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "environment.0.variables.key2", resourceName, "environment.0.variables.key2"),
				),
			},
		},
	})
}

func testAccDataSourceAWSLambdaFunctionConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "lambda" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "lambda" {
  name = %[1]q
  role = "${aws_iam_role.lambda.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
        },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccDataSourceAWSLambdaFunctionConfigBasic(rName string) string {
	return testAccDataSourceAWSLambdaFunctionConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = "${aws_iam_role.lambda.arn}"
  runtime       = "nodejs8.10"
}

data "aws_lambda_function" "test" {
  function_name = "${aws_lambda_function.test.function_name}"
}
`, rName)
}

func testAccDataSourceAWSLambdaFunctionConfigVersion(rName string) string {
	return testAccDataSourceAWSLambdaFunctionConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  publish       = true
  role          = "${aws_iam_role.lambda.arn}"
  runtime       = "nodejs8.10"
}

data "aws_lambda_function" "test" {
  function_name = "${aws_lambda_function.test.function_name}"
  qualifier     = 1
}
`, rName)
}

func testAccDataSourceAWSLambdaFunctionConfigAlias(rName string) string {
	return testAccDataSourceAWSLambdaFunctionConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  publish       = true
  role          = "${aws_iam_role.lambda.arn}"
  runtime       = "nodejs8.10"
}

resource "aws_lambda_alias" "test" {
  name             = "alias-name"
  function_name    = "${aws_lambda_function.test.arn}"
  function_version = "1"
}

data "aws_lambda_function" "test" {
  function_name = "${aws_lambda_function.test.function_name}"
  qualifier     = "${aws_lambda_alias.test.name}"
}
`, rName)
}

func testAccDataSourceAWSLambdaFunctionConfigLayers(rName string) string {
	return testAccDataSourceAWSLambdaFunctionConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs8.10"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  layers        = ["${aws_lambda_layer_version.test.arn}"]
  role          = "${aws_iam_role.lambda.arn}"
  runtime       = "nodejs8.10"
}

data "aws_lambda_function" "test" {
  function_name = "${aws_lambda_function.test.function_name}"
}
`, rName)
}

func testAccDataSourceAWSLambdaFunctionConfigVPC(rName string) string {
	return testAccDataSourceAWSLambdaFunctionConfigBase(rName) + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = "${aws_vpc.test.id}"
  cidr_block = "10.0.1.0/24"
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Terraform Acceptance Testing"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    self        = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = "${aws_iam_role.lambda.arn}"
  runtime       = "nodejs8.10"

  vpc_config {
    security_group_ids = ["${aws_security_group.test.id}"]
    subnet_ids         = ["${aws_subnet.test.id}"]
  }
}

data "aws_lambda_function" "test" {
  function_name = "${aws_lambda_function.test.function_name}"
}
`, rName)
}

func testAccDataSourceAWSLambdaFunctionConfigEnvironment(rName string) string {
	return testAccDataSourceAWSLambdaFunctionConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = "${aws_iam_role.lambda.arn}"
  runtime       = "nodejs8.10"

  environment {
    variables = {
      key1 = "value1"
      key2 = "value2"
    }
  }
}

data "aws_lambda_function" "test" {
  function_name = "${aws_lambda_function.test.function_name}"
}
`, rName)
}
