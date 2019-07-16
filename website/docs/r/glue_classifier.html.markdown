---
layout: "aws"
page_title: "AWS: aws_glue_classifier"
sidebar_current: "docs-aws-resource-glue-classifier"
description: |-
  Provides an Glue Classifier resource.
---

# Resource: aws_glue_classifier

Provides a Glue Classifier resource.

~> **NOTE:** It is only valid to create one type of classifier (grok, JSON, or XML). Changing classifier types will recreate the classifier.

## Example Usage

### Grok Classifier

```hcl
resource "aws_glue_classifier" "example" {
  name = "example"

  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}
```

### JSON Classifier

```hcl
resource "aws_glue_classifier" "example" {
  name = "example"

  json_classifier {
    json_path = "example"
  }
}
```

### XML Classifier

```hcl
resource "aws_glue_classifier" "example" {
  name = "example"

  xml_classifier {
    classification = "example"
    row_tag        = "example"
  }
}
```

## Argument Reference

The following arguments are supported:

* `grok_classifier` – (Optional) A classifier that uses grok patterns. Defined below.
* `json_classifier` – (Optional) A classifier for JSON content. Defined below.
* `name` – (Required) The name of the classifier.
* `xml_classifier` – (Optional) A classifier for XML content. Defined below.

### grok_classifier

* `classification` - (Required) An identifier of the data format that the classifier matches, such as Twitter, JSON, Omniture logs, Amazon CloudWatch Logs, and so on.
* `custom_patterns` - (Optional) Custom grok patterns used by this classifier.
* `grok_pattern` - (Required) The grok pattern used by this classifier.

### json_classifier

* `json_path` - (Required) A `JsonPath` string defining the JSON data for the classifier to classify. AWS Glue supports a subset of `JsonPath`, as described in [Writing JsonPath Custom Classifiers](https://docs.aws.amazon.com/glue/latest/dg/custom-classifier.html#custom-classifier-json).

### xml_classifier

* `classification` - (Required) An identifier of the data format that the classifier matches.
* `row_tag` - (Required) The XML tag designating the element that contains each record in an XML document being parsed. Note that this cannot identify a self-closing element (closed by `/>`). An empty row element that contains only attributes can be parsed as long as it ends with a closing tag (for example, `<row item_a="A" item_b="B"></row>` is okay, but `<row item_a="A" item_b="B" />` is not).

## Attributes Reference

The following additional attributes are exported:

* `id` - Name of the classifier

## Import

Glue Classifiers can be imported using their name, e.g.

```
$ terraform import aws_glue_classifier.MyClassifier MyClassifier
```
