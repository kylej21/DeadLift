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

output "repair_sa_email" {
  description = "Grant this SA access to customer Pub/Sub resources during onboarding"
  value       = google_service_account.repair.email
}

output "vpn_gateway_ip" {
  description = "GCP VPN gateway public IP — configure this as the peer IP on your on-prem VPN device"
  value       = google_compute_address.vpn_ip.address
}
