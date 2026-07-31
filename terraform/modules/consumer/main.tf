resource "kubernetes_deployment" "consumer" {
  metadata {
    name      = "consumer-control-plane"
    namespace = var.namespace
    labels = {
      app = "consumer-control-plane"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "consumer-control-plane"
      }
    }

    template {
      metadata {
        labels = {
          app = "consumer-control-plane"
        }
      }

      spec {
        container {
          name  = "consumer-control-plane"
          image = "dataspace-consumer:${var.image_tag}"

          port {
            container_port = 8091
          }

          env {
            name  = "PORT"
            value = "8091"
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
            value = "consumer-control-plane"
          }
          env {
            name  = "DATABASE_URL"
            value = "postgres://postgres:${var.postgres_password}@postgres:5432/dataspace_identity?sslmode=disable"
          }
          env {
            name  = "PROVIDER_DSP_URL"
            value = "http://control-plane:8081/api/dsp/2025-1"
          }
          env {
            name  = "PROVIDER_ID"
            value = "did:web:provider"
          }
          env {
            name  = "CONSUMER_CALLBACK_URL"
            value = "http://consumer-control-plane:8091/consumer"
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "consumer" {
  metadata {
    name      = "consumer-control-plane"
    namespace = var.namespace
  }

  spec {
    selector = {
      app = "consumer-control-plane"
    }

    port {
      port        = 8091
      target_port = 8091
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
  value = kubernetes_service.consumer.spec[0].port[0].node_port
}

output "cluster_ip" {
  value = kubernetes_service.consumer.spec[0].cluster_ip
}
