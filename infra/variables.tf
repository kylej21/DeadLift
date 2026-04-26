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

variable "graphrag_server_url" {
  description = "Base URL of the GraphRAG onboarding server (host:port)"
  type        = string
  default     = "http://10.30.112.192:2626"
}

variable "onprem_vpn_peer_ip" {
  description = "Public IP of the on-prem VPN peer (your router/firewall external IP)"
  type        = string
}

variable "onprem_vpn_shared_secret" {
  description = "Pre-shared key for the IKEv1/IKEv2 VPN tunnel"
  type        = string
  sensitive   = true
}

variable "onprem_cidr" {
  description = "On-prem network CIDR reachable through the VPN tunnel"
  type        = string
  default     = "10.30.112.0/24"
}
