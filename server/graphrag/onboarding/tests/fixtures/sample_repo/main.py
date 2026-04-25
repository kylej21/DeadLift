"""Entry point: sets up the Pub/Sub subscriber and runs the message loop."""
import yaml
from google.cloud import pubsub_v1

from src.handler import MessageHandler, ValidationError, ProcessingError


def load_config(path: str = "config.yaml") -> dict:
    with open(path) as fh:
        return yaml.safe_load(fh)


def main() -> None:
    cfg = load_config()
    pubsub_cfg = cfg["pubsub"]

    project_id: str = pubsub_cfg["project_id"]           # config_key: pubsub.project_id
    subscription_id: str = pubsub_cfg["subscription"]    # config_key: pubsub.subscription
    dlq_subscription: str = pubsub_cfg["dead_letter_subscription"]  # dead_letter_config

    subscriber = pubsub_v1.SubscriberClient()
    subscription_path = subscriber.subscription_path(project_id, subscription_id)

    handler = MessageHandler()

    def callback(message) -> None:
        try:
            handler.handle(message)
        except ValidationError as exc:
            # Schema validation failure — nack so message routes to dead_letter_config
            print(f"[VALIDATION ERROR] {exc}")
            message.nack()
        except ProcessingError as exc:
            # Transient downstream failure — let retry_policy handle redelivery
            print(f"[PROCESSING ERROR] {exc}")
            message.nack()

    streaming_pull_future = subscriber.subscribe(subscription_path, callback=callback)
    print(f"Listening on {subscription_path} (DLQ: {dlq_subscription})")

    with subscriber:
        try:
            streaming_pull_future.result()
        except KeyboardInterrupt:
            streaming_pull_future.cancel()


if __name__ == "__main__":
    main()
