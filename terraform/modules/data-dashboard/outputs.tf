output "node_port" {
  description = "Data Dashboard NodePort"
  value       = kubernetes_service.data_dashboard.spec[0].port[0].node_port
}
