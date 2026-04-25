output "cloud_run_url" {
  description = "Public URL of the deployed client"
  value       = google_cloud_run_v2_service.client.uri
}

output "artifact_registry_repo" {
  description = "Base image path for the Artifact Registry repo"
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/deadlift"
}

output "wif_provider" {
  description = "Paste this into the WIF_PROVIDER GitHub secret"
  value       = google_iam_workload_identity_pool_provider.github.name
}

output "deployer_service_account" {
  description = "Paste this into the GCP_SA_EMAIL GitHub secret"
  value       = google_service_account.github_actions.email
}

output "proxy_url" {
  description = "URL of the deployed proxy service"
  value       = google_cloud_run_v2_service.proxy.uri
}

output "redirect_uri" {
  description = "Set this as REDIRECT_URI env var on the proxy and in your Google OAuth app"
  value       = "${google_cloud_run_v2_service.proxy.uri}/api/onboard/callback"
}
