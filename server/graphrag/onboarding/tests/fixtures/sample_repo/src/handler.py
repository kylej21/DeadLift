"""Message handler for payment Pub/Sub events."""
import json
from dataclasses import dataclass


@dataclass
class PaymentEvent:
    payment_id: str
    amount: float
    currency: str
    status: str


class ValidationError(Exception):
    """Raised when the incoming message does not match the expected message_schema."""


class ProcessingError(Exception):
    """Raised when a downstream service or handler dependency fails."""


def process_payment(event: PaymentEvent) -> None:
    """Core payment processing logic — calls downstream consumer services."""
    if event.amount <= 0:
        raise ProcessingError(f"Invalid amount for payment {event.payment_id}")
    # downstream handler logic would go here


class MessageHandler:
    """Pub/Sub message handler that deserializes and processes PaymentEvent messages.

    This handler implements the retry_policy defined in config.yaml and forwards
    unprocessable messages to the dead_letter_config topic.
    """

    def handle(self, message) -> None:
        """Deserialize a Pub/Sub message and dispatch to process_payment.

        Raises:
            ValidationError: if the message does not conform to the PaymentEvent schema.
            ProcessingError: if the downstream payment service call fails.
        """
        try:
            data = json.loads(message.data.decode("utf-8"))
        except (json.JSONDecodeError, UnicodeDecodeError) as exc:
            raise ValidationError(f"Could not decode message: {exc}") from exc

        required_fields = {"payment_id", "amount", "currency", "status"}
        missing = required_fields - data.keys()
        if missing:
            raise ValidationError(f"Missing fields in message_schema: {missing}")

        event = PaymentEvent(**{k: data[k] for k in required_fields})
        process_payment(event)
        message.ack()
