variable "namespace" {
  description = "Kubernetes namespace to deploy resources into"
  type        = string
}

variable "postgres_password" {
  description = "Password for the Postgres database"
  type        = string
  sensitive   = true
}
