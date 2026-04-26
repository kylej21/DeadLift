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

# ── Proxy service ─────────────────────────────────────────────────────────────

resource "google_service_account" "proxy" {
  account_id   = "deadlift-proxy"
  display_name = "DeadLift Proxy Service"
}

resource "google_service_account" "repair" {
  account_id   = "deadlift-repair"
  display_name = "DeadLift Repair SA"
  description  = "Granted access to customer Pub/Sub resources during onboarding"
}

resource "google_project_iam_member" "proxy_firestore" {
  project = var.project_id
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.proxy.email}"
}

resource "google_cloud_run_v2_service" "proxy" {
  name     = "deadlift-proxy"
  location = var.region

  template {
    service_account = google_service_account.proxy.email

    scaling {
      min_instance_count = 1
    }

    vpc_access {
      connector = google_vpc_access_connector.connector.id
      egress    = "PRIVATE_RANGES_ONLY"
    }

    containers {
      image = "us-docker.pkg.dev/cloudrun/container/hello"

      ports {
        container_port = 8080
      }

      env {
        name  = "GCP_PROJECT_ID"
        value = var.project_id
      }
      env {
        name  = "GOOGLE_CLIENT_ID"
        value = var.google_oauth_client_id
      }
      env {
        name  = "GOOGLE_CLIENT_SECRET"
        value = var.google_oauth_client_secret
      }
      env {
        name  = "REPAIR_SA_EMAIL"
        value = google_service_account.repair.email
      }
      env {
        name  = "CLIENT_URL"
        value = var.proxy_client_url
      }
      env {
        name  = "REDIRECT_URI"
        value = "https://deadlift-proxy-f47qsb66lq-uc.a.run.app/api/onboard/callback"
      }
      env {
        name  = "GRAPHRAG_SERVER_URL"
        value = var.graphrag_server_url
      }
      env {
        name  = "VLLM_SERVER_URL"
        value = var.vllm_server_url
      }
      env {
        name  = "VLLM_API_KEY"
        value = var.vllm_api_key
      }
      env {
        name  = "VLLM_MODEL"
        value = var.vllm_model
      }
      env {
        name  = "GITHUB_CLIENT_ID"
        value = var.github_client_id
      }
      env {
        name  = "GITHUB_CLIENT_SECRET"
        value = var.github_client_secret
      }
      env {
        name  = "GITHUB_REDIRECT_URI"
        value = var.github_redirect_uri
      }
    }
  }

  lifecycle {
    ignore_changes = [template[0].containers[0].image]
  }

  depends_on = [
    google_project_service.apis,
    google_artifact_registry_repository.client,
    google_firestore_database.default,
  ]
}

resource "google_cloud_run_v2_service_iam_member" "proxy_public" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.proxy.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# REDIRECT_URI must be set after first deploy when the proxy URL is known.
# Update it in the GCP Console or via: gcloud run services update deadlift-proxy --set-env-vars REDIRECT_URI=<proxy_url>/api/onboard/callback
