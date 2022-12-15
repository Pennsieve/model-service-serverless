### SERVICE LAMBDA

## Lambda Function which consumes messages from the SQS queue which contains all events.
resource "aws_lambda_function" "service_lambda" {
  description       = "Lambda Function which handles requests for the serverless metadata model service"
  function_name     = "${var.environment_name}-${var.service_name}-service-lambda-${data.terraform_remote_state.region.outputs.aws_region_shortname}"
  handler           = "model_service"
  runtime           = "go1.x"
  role              = aws_iam_role.model_service_lambda_role.arn
  timeout           = 300
  memory_size       = 128
  s3_bucket         = data.aws_s3_bucket_object.service_lambda_s3_object.bucket
  s3_key            = data.aws_s3_bucket_object.service_lambda_s3_object.key
  s3_object_version = data.aws_s3_bucket_object.service_lambda_s3_object.version_id
#  source_code_hash = data.archive_file.model_service_lambda_archive.output_base64sha256
#  filename         = "${path.module}/../lambda/bin/model_service_handler.zip"

  vpc_config {
    subnet_ids         = tolist(data.terraform_remote_state.vpc.outputs.private_subnet_ids)
    security_group_ids = [data.terraform_remote_state.platform_infrastructure.outputs.upload_v2_security_group_id]
  }

  environment {
    variables = {
      ENV = var.environment_name
      PENNSIEVE_DOMAIN = data.terraform_remote_state.account.outputs.domain_name,
      REGION = var.aws_region
    }
  }
}

#data "archive_file" "model_service_lambda_archive" {
#  type        = "zip"
#  source_dir  = "${path.module}/../lambda/bin/modelService"
#  output_path = "${path.module}/../lambda/bin/model_service_handler.zip"
#}