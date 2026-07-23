variable "kubeconfig_path" {
  description = "Path to the kubeconfig file for cluster connection"
  type        = string
  default     = "~/.kube/config"
}

variable "kubeconfig_context" {
  description = "Context to use in the kubeconfig file"
  type        = string
  default     = "default"
}

variable "namespace" {
  description = "Kubernetes namespace to deploy resources into"
  type        = string
  default     = "sovereign-dataspace"
}

variable "postgres_password" {
  description = "Password for the Postgres database"
  type        = string
  default     = "postgres"
  sensitive   = true
}

variable "image_tag" {
  description = "Docker image tag for all services"
  type        = string
  default     = "latest"
}
