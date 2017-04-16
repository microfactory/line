variable "version" {}
variable "owner" {}
variable "project" {}

variable "region_1" { default = "eu-west-1" }
provider "aws" {
  region = "${var.region_1}"
}

//
// Deployment
//

module "deployment" {
  source = "../../modules/deployment"
  version = "${var.version}"
  owner = "${var.owner}"
  project = "${var.project}"
}

//
// Resources
//

resource "aws_dynamodb_table" "t1" {
  name           = "${module.deployment.id}-GameScores"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserId"
  range_key      = "GameTitle"

  attribute {
    name = "UserId"
    type = "S"
  }

  attribute {
    name = "GameTitle"
    type = "S"
  }

  attribute {
    name = "TopScore"
    type = "N"
  }

  global_secondary_index {
    name               = "idx1"
    hash_key           = "GameTitle"
    range_key          = "TopScore"
    write_capacity     = 1
    read_capacity      = 1
    projection_type    = "INCLUDE"
    non_key_attributes = ["UserId"]
  }
}

//
// Resource Permissions
//

data "aws_iam_policy_document" "observers" {
  policy_id = "${module.deployment.id}-observers"

  statement {
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem",
      "dynamodb:Query",
      "dynamodb:Scan"
    ]
    resources = [
      "${aws_dynamodb_table.t1.arn}*"
    ]
  }
}

//
// Observers
//

module "gateway_observer" {
  source = "../../modules/lambda"
  region = "${var.region_1}"
  deployment = "${module.deployment.id}"

  name = "gateway"
  package = "handler.zip"
  permissions = "${data.aws_iam_policy_document.observers.json}"
  resource_attributes = {
    "my-table-name" = "${aws_dynamodb_table.t1.name}"
    "my-table-name-idx-a" = "${lookup(aws_dynamodb_table.t1.global_secondary_index[0], "name")}"
  }
}

module "schedule_observer" {
  source = "../../modules/lambda"
  region = "${var.region_1}"
  deployment = "${module.deployment.id}"

  name = "schedule"
  package = "handler.zip"
  permissions = "${data.aws_iam_policy_document.observers.json}"
  resource_attributes = {
    "my-table-name" = "${aws_dynamodb_table.t1.name}"
    "my-table-name-idx-a" = "${lookup(aws_dynamodb_table.t1.global_secondary_index[0], "name")}"
  }
}

//
// Streams
//

module "gateway_stream" {
  source = "../../modules/gateway/proxy"
  region = "${var.region_1}"
  deployment = "${module.deployment.id}"

  name = "gateway"
  observer = "${module.gateway_observer.arn}"
}

module "schedule_stream" {
  source = "../../modules/schedule"
  region = "${var.region_1}"
  deployment = "${module.deployment.id}"

  name = "schedule"
  observer = "${module.schedule_observer.arn}"
}
