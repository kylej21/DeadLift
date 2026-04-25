resource "google_cloud_run_v2_service" "client" {
  name     = "deadlift-client"
  location = var.region

  template {
    containers {
      # Placeholder image for initial provisioning; CI overwrites this on first deploy
      image = "us-docker.pkg.dev/cloudrun/container/hello"

      ports {
        container_port = 8080
      }
    }
  }

  lifecycle {
    ignore_changes = [template[0].containers[0].image]
  }

  depends_on = [
    google_project_service.apis,
    google_artifact_registry_repository.client,
  ]
}

# Allow unauthenticated public access
resource "google_cloud_run_v2_service_iam_member" "public" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.client.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
