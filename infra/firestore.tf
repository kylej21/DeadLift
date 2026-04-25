resource "google_firestore_database" "default" {
  name        = "(default)"
  location_id = var.region
  type        = "FIRESTORE_NATIVE"

  depends_on = [google_project_service.apis]
}

# Required for: listTasksByOrg — WHERE org_id == ? ORDER BY created_at DESC
resource "google_firestore_index" "tasks_by_org" {
  collection = "tasks"
  database   = google_firestore_database.default.name

  fields {
    field_path = "org_id"
    order      = "ASCENDING"
  }
  fields {
    field_path = "created_at"
    order      = "DESCENDING"
  }
}
