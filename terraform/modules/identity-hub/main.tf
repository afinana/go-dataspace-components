resource "kubernetes_deployment" "identity_hub" {
  metadata {
    name      = "identity-hub"
    namespace = var.namespace
    labels = {
      app = "identity-hub"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "identity-hub"
      }
    }

    template {
      metadata {
        labels = {
          app = "identity-hub"
        }
      }

      spec {
        container {
          name  = "identity-hub"
          image = "dataspace-identity-hub:${var.image_tag}"

          port {
            container_port = 8080
          }

          env {
            name  = "PORT"
            value = "8080"
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

resource "kubernetes_service" "identity_hub" {
  metadata {
    name      = "identity-hub"
    namespace = var.namespace
  }

  spec {
    selector = {
      app = "identity-hub"
    }

    port {
      port        = 8080
      target_port = 8080
    }
  }
}
