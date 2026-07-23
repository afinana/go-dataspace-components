output "cluster_ip" {
  description = "Identity Hub Cluster IP"
  value       = kubernetes_service.identity_hub.spec[0].cluster_ip
}
