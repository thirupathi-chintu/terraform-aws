---
layout: "aws"
page_title: "AWS: aws_dx_hosted_public_virtual_interface_accepter"
sidebar_current: "docs-aws-resource-dx-hosted-public-virtual-interface-accepter"
description: |-
  Provides a resource to manage the accepter's side of a Direct Connect hosted public virtual interface.
---

# Resource: aws_dx_hosted_public_virtual_interface_accepter

Provides a resource to manage the accepter's side of a Direct Connect hosted public virtual interface.
This resource accepts ownership of a public virtual interface created by another AWS account.

## Example Usage

```hcl
provider "aws" {
  # Creator's credentials.
}

provider "aws" {
  alias = "accepter"

  # Accepter's credentials.
}

data "aws_caller_identity" "accepter" {
  provider = "aws.accepter"
}

# Creator's side of the VIF
resource "aws_dx_hosted_public_virtual_interface" "creator" {
  connection_id    = "dxcon-zzzzzzzz"
  owner_account_id = "${data.aws_caller_identity.accepter.account_id}"

  name           = "vif-foo"
  vlan           = 4094
  address_family = "ipv4"
  bgp_asn        = 65352

  customer_address = "175.45.176.1/30"
  amazon_address   = "175.45.176.2/30"

  route_filter_prefixes = [
    "210.52.109.0/24",
    "175.45.176.0/22",
  ]
}

# Accepter's side of the VIF.
resource "aws_dx_hosted_public_virtual_interface_accepter" "accepter" {
  provider             = "aws.accepter"
  virtual_interface_id = "${aws_dx_hosted_public_virtual_interface.creator.id}"

  tags = {
    Side = "Accepter"
  }
}
```

## Argument Reference

The following arguments are supported:

* `virtual_interface_id` - (Required) The ID of the Direct Connect virtual interface to accept.
* `tags` - (Optional) A mapping of tags to assign to the resource.

### Removing `aws_dx_hosted_public_virtual_interface_accepter` from your configuration

AWS allows a Direct Connect hosted public virtual interface to be deleted from either the allocator's or accepter's side.
However, Terraform only allows the Direct Connect hosted public virtual interface to be deleted from the allocator's side
by removing the corresponding `aws_dx_hosted_public_virtual_interface` resource from your configuration.
Removing a `aws_dx_hosted_public_virtual_interface_accepter` resource from your configuration will remove it
from your statefile and management, **but will not delete the Direct Connect virtual interface.**


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual interface.
* `arn` - The ARN of the virtual interface.

## Timeouts

`aws_dx_hosted_public_virtual_interface_accepter` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for creating virtual interface
- `delete` - (Default `10 minutes`) Used for destroying virtual interface

## Import

Direct Connect hosted public virtual interfaces can be imported using the `vif id`, e.g.

```
$ terraform import aws_dx_hosted_public_virtual_interface_accepter.test dxvif-33cc44dd
```
