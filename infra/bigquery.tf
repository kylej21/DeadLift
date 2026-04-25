resource "google_bigquery_dataset" "deadlift" {
  dataset_id  = "deadlift"
  description = "DeadLift multi-tenant message analytics"
  location    = var.region

  depends_on = [google_project_service.apis]
}

resource "google_bigquery_table" "success_logs" {
  dataset_id          = google_bigquery_dataset.deadlift.dataset_id
  table_id            = "success_logs"
  deletion_protection = false

  # Schema matches what Pub/Sub BigQuery subscriptions write with write_metadata=true.
  # subscription_name encodes the org_id: deadlift-analytics-{org_id}
  schema = jsonencode([
    { name = "subscription_name", type = "STRING",    mode = "NULLABLE" },
    { name = "message_id",        type = "STRING",    mode = "NULLABLE" },
    { name = "publish_time",      type = "TIMESTAMP", mode = "NULLABLE" },
    { name = "data",              type = "STRING",    mode = "NULLABLE" },
    { name = "attributes",        type = "STRING",    mode = "NULLABLE" },
    { name = "ordering_key",      type = "STRING",    mode = "NULLABLE" },
  ])

  time_partitioning {
    type  = "DAY"
    field = "publish_time"
  }

  clustering = ["subscription_name"]
}

# Proxy SA needs bigquery.dataEditor on the dataset so it can
# dynamically add each customer's Pub/Sub service agent during onboarding.
resource "google_bigquery_dataset_iam_member" "proxy_bq_owner" {
  dataset_id = google_bigquery_dataset.deadlift.dataset_id
  role       = "roles/bigquery.dataOwner"
  member     = "serviceAccount:${google_service_account.proxy.email}"
}
