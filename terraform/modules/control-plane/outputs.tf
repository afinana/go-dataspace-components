output "cluster_ip" {
  description = "Control Plane Cluster IP"
  value       = kubernetes_service.control_plane.spec[0].cluster_ip
}
