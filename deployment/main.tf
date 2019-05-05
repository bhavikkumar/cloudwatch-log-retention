terraform {
  backend "s3" {
    key     = "common/lambda/cloudwatch-log-retention"
    encrypt = true
  }
}

locals {
  common_tags = {
    Owner       = "global"
    Environment = "${terraform.workspace}"
  }
}

data "terraform_remote_state" "master" {
  backend = "s3"
  config {
    bucket   = "terraform.bhavik.io"
    key      = "common/master"
    region   = "${var.aws_default_region}"
    profile  = "${var.profile}"
    role_arn = "arn:aws:iam::${var.operations_account_id}:role/${var.role_name}"
  }
}

provider "aws" {
  region  = "${var.aws_default_region}"
  version = "~> 2.8.0"
  profile = "${var.profile}"

  assume_role {
    role_arn     = "arn:aws:iam::${var.account_id}:role/${var.role_name}"
    session_name = "terraform"
  }
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect = "Allow"

    actions = [
      "sts:AssumeRole"
    ]

    principals {
      type = "Service"

      identifiers = [
        "lambda.amazonaws.com"
      ]
    }
  }
}

data "aws_iam_policy_document" "lambda_write_logs" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]

    resources = [
      "${aws_cloudwatch_log_group.lambda.arn}"
    ]
  }
}

data "aws_iam_policy_document" "retention_policy" {
  statement {
    effect = "Allow"

    actions = [
      "logs:PutRetentionPolicy"
    ]

    resources = [
      "arn:aws:logs:*:*:*"
    ]
  }
}

resource "aws_iam_role" "lambda" {
  name               = "CloudWatchRetentionLambda"
  description        = "Used by CloudWatch Retention Lambda"
  assume_role_policy = "${data.aws_iam_policy_document.lambda_assume_role.json}"
  tags               = "${merge(local.common_tags, var.tags)}"
}

resource "aws_iam_role_policy" "lambda_write_logs" {
  name   = "CloudwatchLogWritePermissions"
  role   = "${aws_iam_role.lambda.name}"
  policy = "${data.aws_iam_policy_document.lambda_write_logs.json}"
}

resource "aws_iam_role_policy" "lambda_retention_period_policy" {
  name   = "AllowPutRetentionPeriodPolicy"
  role   = "${aws_iam_role.lambda.name}"
  policy = "${data.aws_iam_policy_document.retention_policy.json}"
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${aws_lambda_function.lambda.function_name}"
  retention_in_days = "${var.log_retention_period}"
  kms_key_id        = "${data.terraform_remote_state.master.default_kms_key_arn}"
  tags              = "${merge(local.common_tags, var.tags)}"
}

resource "aws_cloudwatch_log_subscription_filter" "lambda" {
  name            = "DefaultLogDestination"
  log_group_name  = "${aws_cloudwatch_log_group.lambda.name}"
  filter_pattern  = ""
  destination_arn = "${data.terraform_remote_state.master.log_destination_arn}"
  distribution    = "ByLogStream"
}

resource "aws_lambda_function" "lambda" {
  function_name    = "CloudWatchLogRetention"
  description      = "Sets the default cloudwatch log retention period"
  role             = "${aws_iam_role.lambda.arn}"
  handler          = "main"
  runtime          = "go1.x"
  memory_size      = 128
  kms_key_arn      = "${data.terraform_remote_state.master.default_kms_key_arn}"
  filename         = "cloudwatch-log-retention${var.lambda_version}.zip"
  publish          = true
  source_code_hash = "${filebase64sha256(format("cloudwatch-log-retention%s.zip", var.lambda_version))}"

  environment {
    variables = {
      RETENTION_PERIOD = "${var.log_retention_period}"
    }
  }
  tags = "${merge(local.common_tags, var.tags)}"
}

resource "aws_cloudwatch_event_rule" "retention_period" {
  name        = "LogRetentionPeriodModifications"
  description = "Captures when log groups are created or the retention period is modified"
  tags        = "${merge(local.common_tags, var.tags)}"

  event_pattern = <<PATTERN
{
  "source": [
    "aws.logs"
  ],
  "detail-type": [
    "AWS API Call via CloudTrail"
  ],
  "detail": {
    "eventSource": [
      "logs.amazonaws.com"
    ],
    "eventName": [
      "CreateLogGroup",
      "PutRetentionPolicy",
      "DeleteRetentionPolicy"
    ]
  }
}
PATTERN
}

resource "aws_cloudwatch_event_target" "retention_period_lambda" {
  rule = "${aws_cloudwatch_event_rule.retention_period.name}"
  arn  = "${aws_lambda_function.lambda.arn}"
}

resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id  = "AllowRetentionPeriodLambdaExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.lambda.function_name}"
  principal     = "events.amazonaws.com"
  source_arn    = "${aws_cloudwatch_event_rule.retention_period.arn}"
}
