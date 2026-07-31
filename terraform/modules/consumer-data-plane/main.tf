resource "kubernetes_deployment" "consumer_data_plane" {
  metadata {
    name      = "consumer-data-plane"
    namespace = var.namespace
    labels = {
      app = "consumer-data-plane"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "consumer-data-plane"
      }
    }

    template {
      metadata {
        labels = {
          app = "consumer-data-plane"
        }
      }

      spec {
        container {
          name  = "consumer-data-plane"
          image = "dataspace-consumer-data-plane:${var.image_tag}"

          port {
            container_port = 8092
          }

          env {
            name  = "PORT"
            value = "8092"
          }
          env {
            name  = "LOG_LEVEL"
            value = "DEBUG"
          }
          env {
            name  = "ENVIRONMENT"
            value = "development"
          }
          env {
            name  = "SERVICE_NAME"
            value = "consumer-data-plane"
          }
          env {
            name  = "DATABASE_URL"
            value = "postgres://postgres:${var.postgres_password}@postgres:5432/dataspace_identity?sslmode=disable"
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "consumer_data_plane" {
  metadata {
    name      = "consumer-data-plane"
    namespace = var.namespace
  }

  spec {
    selector = {
      app = "consumer-data-plane"
    }

    port {
      port        = 8092
      target_port = 8092
    }

    type = "NodePort"
  }
}

variable "namespace" {
  type = string
}

variable "image_tag" {
  type = string
}

variable "postgres_password" {
  type = string
}

output "node_port" {
  value = kubernetes_service.consumer_data_plane.spec[0].port[0].node_port
}

output "cluster_ip" {
  value = kubernetes_service.consumer_data_plane.spec[0].cluster_ip
}
