---
layout: "aws"
page_title: "AWS: aws_lb_target_group"
sidebar_current: "docs-aws-resource-elbv2-target-group"
description: |-
  Provides a Target Group resource for use with Load Balancers.
---

# Resource: aws_lb_target_group

Provides a Target Group resource for use with Load Balancer resources.

~> **Note:** `aws_alb_target_group` is known as `aws_lb_target_group`. The functionality is identical.

## Example Usage

### Instance Target Group

```hcl
resource "aws_lb_target_group" "test" {
  name     = "tf-example-lb-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.main.id}"
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}
```

### IP Target Group

```hcl
resource "aws_lb_target_group" "ip-example" {
  name        = "tf-example-lb-tg"
  port        = 80
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = "${aws_vpc.main.id}"
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}
```

### Lambda Target Group

```hcl
resource "aws_lb_target_group" "lambda-example" {
  name        = "tf-example-lb-tg"
  target_type = "lambda"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, Forces new resource) The name of the target group. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`. Cannot be longer than 6 characters.

* `port` - (Optional, Forces new resource) The port on which targets receive traffic, unless overridden when registering a specific target. Required when `target_type` is `instance` or `ip`. Does not apply when `target_type` is `lambda`.
* `protocol` - (Optional, Forces new resource) The protocol to use for routing traffic to the targets. Should be one of "TCP", "TLS", "UDP", "TCP_UDP", "HTTP" or "HTTPS". Required when `target_type` is `instance` or `ip`. Does not apply when `target_type` is `lambda`.
* `vpc_id` - (Optional, Forces new resource) The identifier of the VPC in which to create the target group. Required when `target_type` is `instance` or `ip`. Does not apply when `target_type` is `lambda`.
* `deregistration_delay` - (Optional) The amount time for Elastic Load Balancing to wait before changing the state of a deregistering target from draining to unused. The range is 0-3600 seconds. The default value is 300 seconds.
* `slow_start` - (Optional) The amount time for targets to warm up before the load balancer sends them a full share of requests. The range is 30-900 seconds or 0 to disable. The default value is 0 seconds.
* `lambda_multi_value_headers_enabled` - (Optional) Boolean whether the request and response headers exchanged between the load balancer and the Lambda function include arrays of values or strings. Only applies when `target_type` is `lambda`.
* `proxy_protocol_v2` - (Optional) Boolean to enable / disable support for proxy protocol v2 on Network Load Balancers. See [doc](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html#proxy-protocol) for more information.
* `stickiness` - (Optional) A Stickiness block. Stickiness blocks are documented below. `stickiness` is only valid if used with Load Balancers of type `Application`
* `health_check` - (Optional) A Health Check block. Health Check blocks are documented below.
* `target_type` - (Optional, Forces new resource) The type of target that you must specify when registering targets with this target group.
The possible values are `instance` (targets are specified by instance ID) or `ip` (targets are specified by IP address) or `lambda` (targets are specified by lambda arn).
The default is `instance`. Note that you can't specify targets for a target group using both instance IDs and IP addresses.
If the target type is `ip`, specify IP addresses from the subnets of the virtual private cloud (VPC) for the target group,
the RFC 1918 range (10.0.0.0/8, 172.16.0.0/12, and 192.168.0.0/16), and the RFC 6598 range (100.64.0.0/10).
You can't specify publicly routable IP addresses.
* `tags` - (Optional) A mapping of tags to assign to the resource.

Stickiness Blocks (`stickiness`) support the following:

* `type` - (Required) The type of sticky sessions. The only current possible value is `lb_cookie`.
* `cookie_duration` - (Optional) The time period, in seconds, during which requests from a client should be routed to the same target. After this time period expires, the load balancer-generated cookie is considered stale. The range is 1 second to 1 week (604800 seconds). The default value is 1 day (86400 seconds).
* `enabled` - (Optional) Boolean to enable / disable `stickiness`. Default is `true`

~> **NOTE:** To help facilitate the authoring of modules that support target groups of any protocol, you can define `stickiness` regardless of the protocol chosen. However, for `TCP` target groups, `enabled` must be `false`.

Health Check Blocks (`health_check`):

~> **Note:** The Health Check parameters you can set vary by the `protocol` of
the Target Group. Many parameters cannot be set to custom values for `network`
load balancers at this time. See
http://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_CreateTargetGroup.html
for a complete reference.
Keep in mind, that health checks produce actual requests to the backend.
The underlying function is invoked when `target_type` is set to `lambda`.

* `enabled` - (Optional) Indicates whether  health checks are enabled. Defaults to true.
* `interval` - (Optional) The approximate amount of time, in seconds, between health checks of an individual target. Minimum value 5 seconds, Maximum value 300 seconds. For `lambda` target groups, it needs to be greater as the `timeout` of the underlying `lambda`. Default 30 seconds.
* `path` - (Required for HTTP/HTTPS ALB) The destination for the health check request. Applies to Application Load Balancers only (HTTP/HTTPS), not Network Load Balancers (TCP).
* `port` - (Optional) The port to use to connect with the target. Valid values are either ports 1-65536, or `traffic-port`. Defaults to `traffic-port`.
* `protocol` - (Optional) The protocol to use to connect with the target. Defaults to `HTTP`. Not applicable when `target_type` is `lambda`.
* `timeout` - (Optional) The amount of time, in seconds, during which no response means a failed health check. For Application Load Balancers, the range is 2 to 120 seconds, and the default is 5 seconds for the `instance` target type and 30 seconds for the `lambda` target type. For Network Load Balancers, you cannot set a custom value, and the default is 10 seconds for TCP and HTTPS health checks and 6 seconds for HTTP health checks.
* `healthy_threshold` - (Optional) The number of consecutive health checks successes required before considering an unhealthy target healthy. Defaults to 3.
* `unhealthy_threshold` - (Optional) The number of consecutive health check failures required before considering the target unhealthy . For Network Load Balancers, this value must be the same as the `healthy_threshold`. Defaults to 3.
* `matcher` (Required for HTTP/HTTPS ALB) The HTTP codes to use when checking for a successful response from a target. You can specify multiple values (for example, "200,202") or a range of values (for example, "200-299"). Applies to Application Load Balancers only (HTTP/HTTPS), not Network Load Balancers (TCP).   

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ARN of the Target Group (matches `arn`)
* `arn` - The ARN of the Target Group (matches `id`)
* `arn_suffix` - The ARN suffix for use with CloudWatch Metrics.
* `name` - The name of the Target Group

## Import

Target Groups can be imported using their ARN, e.g.

```
$ terraform import aws_lb_target_group.app_front_end arn:aws:elasticloadbalancing:us-west-2:187416307283:targetgroup/app-front-end/20cfe21448b66314
```
