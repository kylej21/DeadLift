variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "github_repo" {
  description = "GitHub repo in owner/repo format"
  type        = string
  default     = "kylej21/DeadLift"
}

variable "google_oauth_client_id" {
  description = "Google OAuth client ID for the proxy"
  type        = string
  sensitive   = true
}

variable "google_oauth_client_secret" {
  description = "Google OAuth client secret for the proxy"
  type        = string
  sensitive   = true
}

variable "proxy_client_url" {
  description = "URL of the client SPA — proxy redirects here after onboarding"
  type        = string
  default     = "https://deadlift-client-f47qsb66lq-uc.a.run.app"
}
