variable "namespace" {
  type        = string
  description = "Kubernetes namespace"
}

variable "prometheus_host" {
  type        = string
  description = "Hostname for Prometheus Traefik ingress route"
  default     = "promethus.middleland.net"
}

variable "grafana_host" {
  type        = string
  description = "Hostname for Grafana Traefik ingress route"
  default     = "grafana.middleland.net"
}

