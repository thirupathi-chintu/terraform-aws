---
layout: "aws"
page_title: "AWS: aws_msk_cluster"
sidebar_current: "docs-aws-resource-msk-cluster"
description: |-
  Terraform resource for managing an AWS Managed Streaming for Kafka cluster
---

# Resource: aws_msk_cluster

Manages AWS Managed Streaming for Kafka cluster

## Example Usage

```hcl

resource "aws_vpc" "vpc" {
  cidr_block = "192.168.0.0/22"
}

data "aws_availability_zones" "azs" {
  state = "available"
}

resource "aws_subnet" "subnet_az1" {
  availability_zone = "${data.aws_availability_zones.azs.names[0]}"
  cidr_block        = "192.168.0.0/24"
  vpc_id            = "${aws_vpc.vpc.id}"
}

resource "aws_subnet" "subnet_az2" {
  availability_zone = "${data.aws_availability_zones.azs.names[1]}"
  cidr_block        = "192.168.1.0/24"
  vpc_id            = "${aws_vpc.vpc.id}"
}

resource "aws_subnet" "subnet_az3" {
  availability_zone = "${data.aws_availability_zones.azs.names[2]}"
  cidr_block        = "192.168.2.0/24"
  vpc_id            = "${aws_vpc.vpc.id}"
}

resource "aws_security_group" "sg" {
  vpc_id = "${aws_vpc.vpc.id}"
}

resource "aws_kms_key" "kms" {
  description = "example"
}

resource "aws_msk_cluster" "example" {
  cluster_name           = "example"
  kafka_version          = "2.1.0"
  number_of_broker_nodes = 3

  broker_node_group_info {
    instance_type  = "kafka.m5.large"
    ebs_volume_size = "1000"
    client_subnets = [
      "${aws_subnet.subnet_az1.id}",
      "${aws_subnet.subnet_az2.id}",
      "${aws_subnet.subnet_az3.id}",
    ]
    security_groups = [ "${aws_security_group.sg.id}" ]
  }

  encryption_info {
    encryption_at_rest_kms_key_arn = "${aws_kms_key.kms.arn}"
  }

  tags = {
    foo = "bar"
  }
}

output "zookeeper_connect_string" {
  value = "${aws_msk_cluster.example.zookeeper_connect_string}"
}

output "bootstrap_brokers" {
  description = "Plaintext connection host:port pairs"
  value       = "${aws_msk_cluster.example.bootstrap_brokers}"
}

output "bootstrap_brokers_tls" {
  description = "TLS connection host:port pairs"
  value       = "${aws_msk_cluster.example.bootstrap_brokers_tls}"
}
```

## Argument Reference

The following arguments are supported:

* `broker_node_group_info` - (Required) Configuration block for the broker nodes of the Kafka cluster.
* `cluster_name` - (Required) Name of the MSK cluster.
* `kafka_version` - (Required) Specify the desired Kafka software version.
* `number_of_broker_nodes` - (Required) The desired total number of broker nodes in the kafka cluster.  It must be a multiple of the number of specified client subnets.
* `client_authentication` - (Optional) Configuration block for specifying a client authentication. See below.
* `configuration_info` - (Optional) Configuration block for specifying a MSK Configuration to attach to Kafka brokers. See below.
* `encryption_info` - (Optional) Configuration block for specifying encryption. See below.
* `enhanced_monitoring` - (Optional) Specify the desired enhanced MSK CloudWatch monitoring level.  See [Monitoring Amazon MSK with Amazon CloudWatch](https://docs.aws.amazon.com/msk/latest/developerguide/monitoring.html)
* `tags` - (Optional) A mapping of tags to assign to the resource

### broker_node_group_info Argument Reference

* `client_subnets` - (Required) A list of subnets to connect to in client VPC ([documentation](https://docs.aws.amazon.com/msk/1.0/apireference/clusters.html#clusters-prop-brokernodegroupinfo-clientsubnets)).
* `ebs_volume_size` - (Required) The size in GiB of the EBS volume for the data drive on each broker node.
* `instance_type` - (Required) Specify the instance type to use for the kafka brokers. e.g. kafka.m5.large. ([Pricing info](https://aws.amazon.com/msk/pricing/))
* `security_groups` - (Required) A list of the security groups to associate with the elastic network interfaces to control who can communicate with the cluster.
* `az_distribution` - (Optional) The distribution of broker nodes across availability zones ([documentation](https://docs.aws.amazon.com/msk/1.0/apireference/clusters.html#clusters-model-brokerazdistribution)). Currently the only valid value is `DEFAULT`.

### client_authentication Argument Reference

* `tls` - (Optional) Configuration block for specifying TLS client authentication. See below.

#### client_authenication tls Argument Reference

* `certificate_authority_arns` - (Optional) List of ACM Certificate Authority Amazon Resource Names (ARNs).

### configuration_info Argument Reference

* `arn` - (Required) Amazon Resource Name (ARN) of the MSK Configuration to use in the cluster.
* `revision` - (Required) Revision of the MSK Configuration to use in the cluster.

### encryption_info Argument Reference

* `encryption_in_transit` - (Optional) Configuration block to specify encryption in transit. See below.
* `encryption_at_rest_kms_key_arn` - (Optional) You may specify a KMS key short ID or ARN (it will always output an ARN) to use for encrypting your data at rest.  If no key is specified, an AWS managed KMS ('aws/msk' managed service) key will be used for encrypting the data at rest.

#### encryption_info encryption_in_transit Argument Reference

* `client_broker` - (Optional) Encryption setting for data in transit between clients and brokers. Valid values: `TLS`, `TLS_PLAINTEXT`, and `PLAINTEXT`. Default value: `TLS_PLAINTEXT`.
* `in_cluster` - (Optional) Whether data communication among broker nodes is encrypted. Default value: `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the MSK cluster.
* `bootstrap_brokers` - A comma separated list of one or more hostname:port pairs of kafka brokers suitable to boostrap connectivity to the kafka cluster. Only contains value if `client_broker` encryption in transit is set to `PLAINTEXT` or `TLS_PLAINTEXT`.
* `bootstrap_brokers_tls` - A comma separated list of one or more DNS names (or IPs) and TLS port pairs kafka brokers suitable to boostrap connectivity to the kafka cluster. Only contains value if `client_broker` encryption in transit is set to `TLS_PLAINTEXT` or `TLS`.
* `current_version` - Current version of the MSK Cluster used for updates, e.g. `K13V1IB3VIYZZH`
* `encryption_info.0.encryption_at_rest_kms_key_arn` - The ARN of the KMS key used for encryption at rest of the broker data volumes.
* `zookeeper_connect_string` - A comma separated list of one or more IP:port pairs to use to connect to the Apache Zookeeper cluster.

## Import

MSK clusters can be imported using the cluster `arn`, e.g.

```
$ terraform import aws_msk_cluster.example arn:aws:kafka:us-west-2:123456789012:cluster/example/279c0212-d057-4dba-9aa9-1c4e5a25bfc7-3
```
