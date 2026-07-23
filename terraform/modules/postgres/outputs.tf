output "cluster_ip" {
  description = "Postgres Cluster IP"
  value       = kubernetes_service.postgres.spec[0].cluster_ip
}
