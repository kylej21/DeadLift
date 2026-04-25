#!/usr/bin/env bash
# Deletes all documents from the Firestore users collection.
# IAM bindings don't need to be revoked — re-onboarding just re-applies them.
#
# Usage: ./scripts/cleanup-users.sh [project_id]

set -euo pipefail

PROJECT_ID="${1:-pubsub-demo-kyle-johnson-2026}"
TOKEN=$(gcloud auth print-access-token)
BASE="https://firestore.googleapis.com/v1/projects/${PROJECT_ID}/databases/(default)/documents"

echo "Fetching users from Firestore project: ${PROJECT_ID}"

docs=$(curl -sf -H "Authorization: Bearer ${TOKEN}" "${BASE}/users" \
  | python3 -c "import sys,json; docs=json.load(sys.stdin).get('documents',[]); [print(d['name']) for d in docs]" 2>/dev/null || true)

if [[ -z "$docs" ]]; then
  echo "No user documents found."
  exit 0
fi

while IFS= read -r doc; do
  echo "Deleting: ${doc##*/}"
  curl -sf -X DELETE -H "Authorization: Bearer ${TOKEN}" \
    "https://firestore.googleapis.com/v1/${doc}" > /dev/null
done <<< "$docs"

echo "Done."
