output "prometheus_service_name" {
  value       = kubernetes_service.prometheus.metadata[0].name
  description = "Prometheus service name"
}

output "grafana_service_name" {
  value       = kubernetes_service.grafana.metadata[0].name
  description = "Grafana service name"
}
