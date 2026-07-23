output "cluster_ip" {
  description = "Data Plane Cluster IP"
  value       = kubernetes_service.data_plane.spec[0].cluster_ip
}
