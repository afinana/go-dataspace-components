output "dashboard_service_url" {
  description = "Access URL for the Data Dashboard"
  value       = "http://localhost:${module.data_dashboard.node_port}"
}

output "postgres_service_cluster_ip" {
  description = "Postgres Cluster IP"
  value       = module.postgres.cluster_ip
}

output "identity_hub_service_cluster_ip" {
  description = "Identity Hub Cluster IP"
  value       = module.identity_hub.cluster_ip
}

output "control_plane_service_cluster_ip" {
  description = "Control Plane Cluster IP"
  value       = module.control_plane.cluster_ip
}

output "data_plane_service_cluster_ip" {
  description = "Data Plane Cluster IP"
  value       = module.data_plane.cluster_ip
}
