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

  schema = jsonencode([
    { name = "org_id",     type = "STRING",    mode = "REQUIRED" },
    { name = "message_id", type = "STRING",    mode = "NULLABLE" },
    { name = "timestamp",  type = "TIMESTAMP", mode = "REQUIRED" },
    { name = "event_type", type = "STRING",    mode = "NULLABLE" },
    { name = "status",     type = "STRING",    mode = "NULLABLE" },
    { name = "payload",    type = "JSON",      mode = "NULLABLE" },
    { name = "attributes", type = "JSON",      mode = "NULLABLE" },
  ])

  time_partitioning {
    type  = "DAY"
    field = "timestamp"
  }

  clustering = ["org_id", "event_type"]
}
