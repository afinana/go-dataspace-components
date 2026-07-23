resource "kubernetes_deployment" "control_plane" {
  metadata {
    name      = "control-plane"
    namespace = var.namespace
    labels = {
      app = "control-plane"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "control-plane"
      }
    }

    template {
      metadata {
        labels = {
          app = "control-plane"
        }
      }

      spec {
        container {
          name  = "control-plane"
          image = "dataspace-control-plane:${var.image_tag}"

          port {
            container_port = 8081
          }

          env {
            name  = "PORT"
            value = "8081"
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

resource "kubernetes_service" "control_plane" {
  metadata {
    name      = "control-plane"
    namespace = var.namespace
  }

  spec {
    selector = {
      app = "control-plane"
    }

    port {
      port        = 8081
      target_port = 8081
    }
  }
}
