---
layout: "aws"
page_title: "AWS: aws_docdb_cluster_parameter_group"
sidebar_current: "docs-aws-resource-docdb-cluster-parameter-group"
description: |-
  Manages a DocumentDB Cluster Parameter Group
---

# Resource: aws_docdb_cluster_parameter_group

Manages a DocumentDB Cluster Parameter Group

## Example Usage

```hcl
resource "aws_docdb_cluster_parameter_group" "example" {
  family      = "docdb3.6"
  name        = "example"
  description = "docdb cluster parameter group"

  parameter {
    name  = "tls"
    value = "enabled"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, Forces new resource) The name of the documentDB cluster parameter group. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `family` - (Required, Forces new resource) The family of the documentDB cluster parameter group.
* `description` - (Optional, Forces new resource) The description of the documentDB cluster parameter group. Defaults to "Managed by Terraform".
* `parameter` - (Optional) A list of documentDB parameters to apply.
* `tags` - (Optional) A mapping of tags to assign to the resource.

Parameter blocks support the following:

* `name` - (Required) The name of the documentDB parameter.
* `value` - (Required) The value of the documentDB parameter.
* `apply_method` - (Optional) Valid values are `immediate` and `pending-reboot`. Defaults to `pending-reboot`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The documentDB cluster parameter group name.
* `arn` - The ARN of the documentDB cluster parameter group.


## Import

DocumentDB Cluster Parameter Groups can be imported using the `name`, e.g.

```
$ terraform import aws_docdb_cluster_parameter_group.cluster_pg production-pg-1
```
