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

resource "kubernetes_manifest" "data_dashboard_ingressroute" {
  manifest = {
    apiVersion = "traefik.io/v1alpha1"
    kind       = "IngressRoute"
    metadata = {
      name      = "data-dashboard-ingressroute"
      namespace = var.namespace
    }
    spec = {
      entryPoints = ["web"]
      routes = [
        {
          match = "Host(`${var.host}`)"
          kind  = "Rule"
          services = [
            {
              name = kubernetes_service.data_dashboard.metadata[0].name
              port = 8084
            }
          ]
        }
      ]
    }
  }
}

