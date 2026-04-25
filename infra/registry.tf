resource "google_artifact_registry_repository" "client" {
  location      = var.region
  repository_id = "deadlift"
  description   = "Docker images for DeadLift client"
  format        = "DOCKER"

  depends_on = [google_project_service.apis]
}
