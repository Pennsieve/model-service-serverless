##############################
# MODEL-SERVICE-LAMBDA   #
##############################

resource "aws_iam_role" "model_service_lambda_role" {
  name = "${var.environment_name}-${var.service_name}-model_service-lambda-role-${data.terraform_remote_state.region.outputs.aws_region_shortname}"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "model_service_lambda_iam_policy_attachment" {
  role       = aws_iam_role.model_service_lambda_role.name
  policy_arn = aws_iam_policy.model_service_lambda_iam_policy.arn
}

resource "aws_iam_policy" "model_service_lambda_iam_policy" {
  name   = "${var.environment_name}-${var.service_name}-model-service-lambda-iam-policy-${data.terraform_remote_state.region.outputs.aws_region_shortname}"
  path   = "/"
  policy = data.aws_iam_policy_document.model_service_iam_policy_document.json
}

data "aws_iam_policy_document" "model_service_iam_policy_document" {

  statement {
    sid    = "SecretsManagerPermissions"
    effect = "Allow"

    actions = [
      "kms:Decrypt",
      "secretsmanager:GetSecretValue",
    ]

    resources = [
      data.aws_kms_key.ssm_kms_key.arn,
    ]
  }

  statement {
    sid    = "ModelServiceLambdaPermissions"
    effect = "Allow"
    actions = [
      "rds-db:connect",
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutDestination",
      "logs:PutLogEvents",
      "logs:DescribeLogStreams",
      "ec2:CreateNetworkInterface",
      "ec2:DescribeNetworkInterfaces",
      "ec2:DeleteNetworkInterface",
      "ec2:AssignPrivateIpAddresses",
      "ec2:UnassignPrivateIpAddresses"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "SSMPermissions"
    effect = "Allow"

    actions = [
      "ssm:GetParameter",
      "ssm:GetParameters",
      "ssm:GetParametersByPath",
    ]

    resources = ["arn:aws:ssm:${data.aws_region.current_region.name}:${data.aws_caller_identity.current.account_id}:parameter/${var.environment_name}/${var.service_name}/*"]
  }
}



