variable "namespace" {
  description = "Kubernetes namespace to deploy resources into"
  type        = string
}

variable "image_tag" {
  description = "Docker image tag for the service"
  type        = string
}

variable "host" {
  description = "Hostname for Traefik ingress route"
  type        = string
  default     = "portal.middleland.net"
}

