// Create log group for upload lambda.
resource "aws_cloudwatch_log_group" "model_service_lambda_loggroup" {
  name              = "/aws/lambda/${aws_lambda_function.service_lambda.function_name}"
  retention_in_days = 30
  tags              = local.common_tags
}
