---
layout: "aws"
page_title: "AWS: aws_securityhub_account"
sidebar_current: "docs-aws-resource-securityhub-account"
description: |-
  Enables Security Hub for an AWS account.
---

# Resource: aws_securityhub_account

Enables Security Hub for this AWS account.

~> **NOTE:** Destroying this resource will disable Security Hub for this AWS account.

~> **NOTE:** This AWS service is in Preview and may change before General Availability release. Backwards compatibility is not guaranteed between Terraform AWS Provider releases.

## Example Usage

```hcl
resource "aws_securityhub_account" "example" {}
```

## Argument Reference

The resource does not support any arguments.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - AWS Account ID.

## Import

An existing Security Hub enabled account can be imported using the AWS account ID, e.g.

```
$ terraform import aws_securityhub_account.example 123456789012
```
