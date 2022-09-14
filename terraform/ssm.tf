resource "aws_ssm_parameter" "neo4j_uri" {
  name  = "/${var.environment_name}/${var.service_name}/db-host"
  type  = "String"
  value = var.neo4j_bolt_url

}

resource "aws_ssm_parameter" "neo4j_user" {
  name  = "/${var.environment_name}/${var.service_name}/neo4j-bolt-user"
  type  = "String"
  value = var.neo4j_bolt_user
}

resource "aws_ssm_parameter" "neo4j_password" {
  name  = "/${var.environment_name}/${var.service_name}/neo4j-bolt-password"
  overwrite = false
  type  = "SecureString"
  value = "dummy"

  lifecycle {
    ignore_changes = [value]
  }
}