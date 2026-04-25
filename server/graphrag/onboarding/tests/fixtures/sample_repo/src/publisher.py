"""Publisher for payment events to the Pub/Sub topic."""
import json
from typing import List

from google.cloud import pubsub_v1

from handler import PaymentEvent


class PaymentPublisher:
    """Publishes PaymentEvent messages to the payment-events topic.

    The topic and project_id are sourced from config.yaml (config_key: pubsub.topic).
    """

    def __init__(self, project_id: str, topic: str) -> None:
        self._client = pubsub_v1.PublisherClient()
        self._topic_path = self._client.topic_path(project_id, topic)

    def publish(self, event: PaymentEvent) -> str:
        """Serialize and publish a single PaymentEvent; returns the message ID."""
        data = json.dumps(
            {
                "payment_id": event.payment_id,
                "amount": event.amount,
                "currency": event.currency,
                "status": event.status,
            }
        ).encode("utf-8")
        future = self._client.publish(self._topic_path, data)
        return future.result()

    def publish_batch(self, events: List[PaymentEvent]) -> List[str]:
        """Publish a batch of PaymentEvent messages; returns list of message IDs."""
        return [self.publish(event) for event in events]
