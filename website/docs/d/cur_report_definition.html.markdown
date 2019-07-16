---
layout: "aws"
page_title: "AWS: aws_cur_report_definition"
sidebar_current: "docs-aws-datasource-cur-report-definition"
description: |-
  Get information on an AWS Cost and Usage Report Definition.
---

# Data Source: aws_cur_report_definition

Use this data source to get information on an AWS Cost and Usage Report Definition.

~> *NOTE:* The AWS Cost and Usage Report service is only available in `us-east-1` currently.

~> *NOTE:* If AWS Organizations is enabled, only the master account can use this resource.

## Example Usage

```hcl
data "aws_cur_report_definition" "report_definition" {
  report_name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `report_name` - (Required) The name of the report definition to match.

## Attributes Reference

* `time_unit` - The frequency on which report data are measured and displayed.
* `format` - Preferred compression format for report.
* `compression` - Preferred format for report.
* `additional_schema_elements` - A list of schema elements.
* `s3_bucket` - Name of customer S3 bucket.
* `s3_prefix` - Preferred report path prefix.
* `s3_region` - Region of customer S3 bucket.
* `additional_artifacts` - A list of additional artifacts.


