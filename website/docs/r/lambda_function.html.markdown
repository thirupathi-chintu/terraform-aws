---
layout: "aws"
page_title: "AWS: aws_lambda_function"
sidebar_current: "docs-aws-resource-lambda-function"
description: |-
  Provides a Lambda Function resource. Lambda allows you to trigger execution of code in response to events in AWS. The Lambda Function itself includes source code and runtime configuration.
---

# Resource: aws_lambda_function

Provides a Lambda Function resource. Lambda allows you to trigger execution of code in response to events in AWS. The Lambda Function itself includes source code and runtime configuration.

For information about Lambda and how to use it, see [What is AWS Lambda?][1]

## Example Usage

### Basic Example

```hcl
resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"

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

resource "aws_lambda_function" "test_lambda" {
  filename      = "lambda_function_payload.zip"
  function_name = "lambda_function_name"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  handler       = "exports.test"

  # The filebase64sha256() function is available in Terraform 0.11.12 and later
  # For Terraform 0.11.11 and earlier, use the base64sha256() function and the file() function:
  # source_code_hash = "${base64sha256(file("lambda_function_payload.zip"))}"
  source_code_hash = "${filebase64sha256("lambda_function_payload.zip")}"

  runtime = "nodejs8.10"

  environment {
    variables = {
      foo = "bar"
    }
  }
}
```

### Lambda Layers

~> **NOTE:** The `aws_lambda_layer_version` attribute values for `arn` and `layer_arn` were swapped in version 2.0.0 of the Terraform AWS Provider. For version 1.x, use `layer_arn` references. For version 2.x, use `arn` references.

```hcl
resource "aws_lambda_layer_version" "example" {
  # ... other configuration ...
}

resource "aws_lambda_function" "example" {
  # ... other configuration ...
  layers = ["${aws_lambda_layer_version.example.arn}"]
}
```

## CloudWatch Logging and Permissions

For more information about CloudWatch Logs for Lambda, see the [Lambda User Guide](https://docs.aws.amazon.com/lambda/latest/dg/monitoring-functions-logs.html).

```hcl
resource "aws_lambda_function" "test_lambda" {
  function_name = "${var.lambda_function_name}"
  ...
  depends_on    = ["aws_iam_role_policy_attachment.lambda_logs", "aws_cloudwatch_log_group.example"]
}

# This is to optionally manage the CloudWatch Log Group for the Lambda Function.
# If skipping this resource configuration, also add "logs:CreateLogGroup" to the IAM policy below.
resource "aws_cloudwatch_log_group" "example" {
  name              = "/aws/lambda/${var.lambda_function_name}"
  retention_in_days = 14
}

# See also the following AWS managed policy: AWSLambdaBasicExecutionRole
resource "aws_iam_policy" "lambda_logging" {
  name = "lambda_logging"
  path = "/"
  description = "IAM policy for logging from a lambda"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role = "${aws_iam_role.iam_for_lambda.name}"
  policy_arn = "${aws_iam_policy.lambda_logging.arn}"
}
```

## Specifying the Deployment Package

AWS Lambda expects source code to be provided as a deployment package whose structure varies depending on which `runtime` is in use.
See [Runtimes][6] for the valid values of `runtime`. The expected structure of the deployment package can be found in
[the AWS Lambda documentation for each runtime][8].

Once you have created your deployment package you can specify it either directly as a local file (using the `filename` argument) or
indirectly via Amazon S3 (using the `s3_bucket`, `s3_key` and `s3_object_version` arguments). When providing the deployment
package via S3 it may be useful to use [the `aws_s3_bucket_object` resource](s3_bucket_object.html) to upload it.

For larger deployment packages it is recommended by Amazon to upload via S3, since the S3 API has better support for uploading
large files efficiently.

## Argument Reference

* `filename` - (Optional) The path to the function's deployment package within the local filesystem. If defined, The `s3_`-prefixed options cannot be used.
* `s3_bucket` - (Optional) The S3 bucket location containing the function's deployment package. Conflicts with `filename`. This bucket must reside in the same AWS region where you are creating the Lambda function.
* `s3_key` - (Optional) The S3 key of an object containing the function's deployment package. Conflicts with `filename`.
* `s3_object_version` - (Optional) The object version containing the function's deployment package. Conflicts with `filename`.
* `function_name` - (Required) A unique name for your Lambda Function.
* `dead_letter_config` - (Optional) Nested block to configure the function's *dead letter queue*. See details below.
* `handler` - (Required) The function [entrypoint][3] in your code.
* `role` - (Required) IAM role attached to the Lambda Function. This governs both who / what can invoke your Lambda Function, as well as what resources our Lambda Function has access to. See [Lambda Permission Model][4] for more details.
* `description` - (Optional) Description of what your Lambda Function does.
* `layers` - (Optional) List of Lambda Layer Version ARNs (maximum of 5) to attach to your Lambda Function. See [Lambda Layers][10]
* `memory_size` - (Optional) Amount of memory in MB your Lambda Function can use at runtime. Defaults to `128`. See [Limits][5]
* `runtime` - (Required) See [Runtimes][6] for valid values.
* `timeout` - (Optional) The amount of time your Lambda Function has to run in seconds. Defaults to `3`. See [Limits][5]
* `reserved_concurrent_executions` - (Optional) The amount of reserved concurrent executions for this lambda function. A value of `0` disables lambda from being triggered and `-1` removes any concurrency limitations. Defaults to Unreserved Concurrency Limits `-1`. See [Managing Concurrency][9]
* `publish` - (Optional) Whether to publish creation/change as new Lambda Function Version. Defaults to `false`.
* `vpc_config` - (Optional) Provide this to allow your function to access your VPC. Fields documented below. See [Lambda in VPC][7]
* `environment` - (Optional) The Lambda environment's configuration settings. Fields documented below.
* `kms_key_arn` - (Optional) The ARN for the KMS encryption key.
* `source_code_hash` - (Optional) Used to trigger updates. Must be set to a base64-encoded SHA256 hash of the package file specified with either `filename` or `s3_key`. The usual way to set this is `filebase64sha256("file.zip")` (Terraform 0.11.12 and later) or `base64sha256(file("file.zip"))` (Terraform 0.11.11 and earlier), where "file.zip" is the local filename of the lambda function source archive.
* `tags` - (Optional) A mapping of tags to assign to the object.

**dead_letter_config** is a child block with a single argument:

* `target_arn` - (Required) The ARN of an SNS topic or SQS queue to notify when an invocation fails. If this
  option is used, the function's IAM role must be granted suitable access to write to the target object,
  which means allowing either the `sns:Publish` or `sqs:SendMessage` action on this ARN, depending on
  which service is targeted.

**tracing_config** is a child block with a single argument:

* `mode` - (Required) Can be either `PassThrough` or `Active`. If PassThrough, Lambda will only trace
  the request from an upstream service if it contains a tracing header with
  "sampled=1". If Active, Lambda will respect any tracing header it receives
  from an upstream service. If no tracing header is received, Lambda will call
  X-Ray for a tracing decision.

**vpc_config** requires the following:

* `subnet_ids` - (Required) A list of subnet IDs associated with the Lambda function.
* `security_group_ids` - (Required) A list of security group IDs associated with the Lambda function.

~> **NOTE:** if both `subnet_ids` and `security_group_ids` are empty then vpc_config is considered to be empty or unset.

For **environment** the following attributes are supported:

* `variables` - (Optional) A map that defines environment variables for the Lambda function.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) identifying your Lambda Function.
* `qualified_arn` - The Amazon Resource Name (ARN) identifying your Lambda Function Version
  (if versioning is enabled via `publish = true`).
* `invoke_arn` - The ARN to be used for invoking Lambda Function from API Gateway - to be used in [`aws_api_gateway_integration`](/docs/providers/aws/r/api_gateway_integration.html)'s `uri`
* `version` - Latest published version of your Lambda Function.
* `last_modified` - The date this resource was last modified.
* `kms_key_arn` - (Optional) The ARN for the KMS encryption key.
* `source_code_hash` - Base64-encoded representation of raw SHA-256 sum of the zip file, provided either via `filename` or `s3_*` parameters.
* `source_code_size` - The size in bytes of the function .zip file.

[1]: https://docs.aws.amazon.com/lambda/latest/dg/welcome.html
[2]: https://docs.aws.amazon.com/lambda/latest/dg/walkthrough-s3-events-adminuser-create-test-function-create-function.html
[3]: https://docs.aws.amazon.com/lambda/latest/dg/walkthrough-custom-events-create-test-function.html
[4]: https://docs.aws.amazon.com/lambda/latest/dg/intro-permission-model.html
[5]: https://docs.aws.amazon.com/lambda/latest/dg/limits.html
[6]: https://docs.aws.amazon.com/lambda/latest/dg/API_CreateFunction.html#SSS-CreateFunction-request-Runtime
[7]: http://docs.aws.amazon.com/lambda/latest/dg/vpc.html
[8]: https://docs.aws.amazon.com/lambda/latest/dg/deployment-package-v2.html
[9]: https://docs.aws.amazon.com/lambda/latest/dg/concurrent-executions.html
[10]: https://docs.aws.amazon.com/lambda/latest/dg/configuration-layers.html

## Timeouts

`aws_lambda_function` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

* `create` - (Default `10m`) How long to wait for slow uploads or EC2 throttling errors.

## Import

Lambda Functions can be imported using the `function_name`, e.g.

```
$ terraform import aws_lambda_function.test_lambda my_test_lambda_function
```
