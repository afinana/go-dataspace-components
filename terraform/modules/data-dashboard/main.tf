resource "kubernetes_deployment" "data_dashboard" {
  metadata {
    name      = "data-dashboard"
    namespace = var.namespace
    labels = {
      app = "data-dashboard"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "data-dashboard"
      }
    }

    template {
      metadata {
        labels = {
          app = "data-dashboard"
        }
      }

      spec {
        container {
          name  = "data-dashboard"
          image = "dataspace-data-dashboard:${var.image_tag}"

          port {
            container_port = 8084
          }

          env {
            name  = "PORT"
            value = "8084"
          }

          env {
            name  = "LOG_LEVEL"
            value = "DEBUG"
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "data_dashboard" {
  metadata {
    name      = "data-dashboard"
    namespace = var.namespace
  }

  spec {
    type = "NodePort"

    selector = {
      app = "data-dashboard"
    }

    port {
      port        = 8084
      target_port = 8084
      node_port   = 30084
    }
  }
}
