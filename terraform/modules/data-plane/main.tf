resource "kubernetes_deployment" "data_plane" {
  metadata {
    name      = "data-plane"
    namespace = var.namespace
    labels = {
      app = "data-plane"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "data-plane"
      }
    }

    template {
      metadata {
        labels = {
          app = "data-plane"
        }
      }

      spec {
        container {
          name  = "data-plane"
          image = "dataspace-data-plane:${var.image_tag}"

          port {
            container_port = 8082
          }

          env {
            name  = "PORT"
            value = "8082"
          }

          env {
            name  = "LOG_LEVEL"
            value = "DEBUG"
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

resource "kubernetes_service" "data_plane" {
  metadata {
    name      = "data-plane"
    namespace = var.namespace
  }

  spec {
    selector = {
      app = "data-plane"
    }

    port {
      port        = 8082
      target_port = 8082
    }
  }
}
