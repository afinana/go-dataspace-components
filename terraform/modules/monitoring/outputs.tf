output "prometheus_service_name" {
  value       = kubernetes_service.prometheus.metadata[0].name
  description = "Prometheus service name"
}

output "grafana_service_name" {
  value       = kubernetes_service.grafana.metadata[0].name
  description = "Grafana service name"
}

output "prometheus_ingress_host" {
  value       = var.prometheus_host
  description = "Prometheus Ingress Host"
}

output "grafana_ingress_host" {
  value       = var.grafana_host
  description = "Grafana Ingress Host"
}

