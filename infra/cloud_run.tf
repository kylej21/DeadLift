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
        value = var.repair_sa_email
      }
      env {
        name  = "CLIENT_URL"
        value = var.proxy_client_url
      }
    }
  }

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,
      template[0].containers[0].env,
    ]
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
