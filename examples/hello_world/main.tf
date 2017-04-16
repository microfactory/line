variable "version" {}
variable "owner" {}
variable "project" {}

variable "region_1" { default = "eu-west-1" }

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
// Service Resources
//

resource "aws_dynamodb_table" "basic-dynamodb-table" {
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
}

//
// Permission Policy
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
      "${aws_dynamodb_table.basic-dynamodb-table.arn}*"
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
  implementation = "handler.zip"
  policy = "${data.aws_iam_policy_document.observers.json}"
  env = {
    "LINE_STRIP_BASEPATH" = 1
  }
}

module "schedule_observer" {
  source = "../../modules/lambda"
  region = "${var.region_1}"
  deployment = "${module.deployment.id}"

  name = "schedule"
  implementation = "handler.zip"
  policy = "${data.aws_iam_policy_document.observers.json}"
  env = {
    "FOO" = "BAR"
  }
}

//
// Streams
//

module "gateway_stream" {
  source = "../../modules/gateway"
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
