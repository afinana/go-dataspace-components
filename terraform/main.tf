terraform {
  required_version = ">= 1.0"
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.kubeconfig_context
}

# Namespace
resource "kubernetes_namespace" "dataspace" {
  metadata {
    name = var.namespace
  }
}

module "postgres" {
  source            = "./modules/postgres"
  namespace         = kubernetes_namespace.dataspace.metadata[0].name
  postgres_password = var.postgres_password
}

module "identity_hub" {
  source            = "./modules/identity-hub"
  namespace         = kubernetes_namespace.dataspace.metadata[0].name
  image_tag         = var.image_tag
  postgres_password = var.postgres_password
}

module "control_plane" {
  source            = "./modules/control-plane"
  namespace         = kubernetes_namespace.dataspace.metadata[0].name
  image_tag         = var.image_tag
  postgres_password = var.postgres_password
}

module "data_plane" {
  source            = "./modules/data-plane"
  namespace         = kubernetes_namespace.dataspace.metadata[0].name
  image_tag         = var.image_tag
  postgres_password = var.postgres_password
}

module "data_dashboard" {
  source    = "./modules/data-dashboard"
  namespace = kubernetes_namespace.dataspace.metadata[0].name
  image_tag = var.image_tag
}
