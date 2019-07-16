---
layout: "aws"
page_title: "AWS: aws_simpledb_domain"
sidebar_current: "docs-aws-resource-simpledb-domain"
description: |-
  Provides a SimpleDB domain resource.
---

# Resource: aws_simpledb_domain

Provides a SimpleDB domain resource

## Example Usage

```hcl
resource "aws_simpledb_domain" "users" {
  name = "users"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the SimpleDB domain

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the SimpleDB domain

## Import

SimpleDB Domains can be imported using the `name`, e.g.

```
$ terraform import aws_simpledb_domain.users users
```
