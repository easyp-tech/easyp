<!-- scope: **/events/*, **/messaging/*, **/pubsub/*, **/subscribers/*, **/producers/*, **/consumers/* -->
# Events Documentation Template

## What This Generates

- `.spec/EVENTS.md` — event catalog, schemas, producers/consumers, delivery guarantees, error handling

## Instructions

You are a technical documentarian. Create event-driven architecture documentation for the project in the `.spec/` directory.
Analyze: event definitions, message schemas, producer/publisher implementations, consumer/subscriber handlers, queue/topic configuration, and retry/DLQ logic.

### Step 1: Identify Event System

Check for the presence of:
- **Message Brokers**: Kafka, RabbitMQ, NATS, Redis Pub/Sub, AWS SQS/SNS, Google Pub/Sub, Azure Service Bus
- **In-Process Event Bus**: EventEmitter (Node.js), mediator pattern, domain events (DDD), Spring Events
- **Streaming**: Kafka Streams, NATS JetStream, Redis Streams, Flink
- **Browser/Frontend**: CustomEvent, EventTarget, RxJS Subject, Vue EventBus, React event callbacks
- **Mobile**: Flutter Streams, BroadcastChannel, NotificationCenter (iOS), LocalBroadcastManager (Android)
- **Webhooks**: outbound event delivery over HTTP (see also `webhooks.md` if you create a separate template)

For each system found, identify: client library, connection config, serialization format (JSON, Protobuf, Avro, MessagePack).

### Step 2: Create .spec/EVENTS.md

#### Structure:

##### 1. Overview
- One sentence: which event system(s) and why
- Architecture pattern: event-driven, CQRS, event sourcing, pub/sub, saga/choreography
- ASCII diagram of the event flow:

```
Producer A ──→ [Topic/Queue] ──→ Consumer X
Producer B ──→ [Topic/Queue] ──→ Consumer Y
                    ↓
              [Dead Letter Queue]
```

##### 2. Event Catalog

Master table of all events in the system:

| Event Name | Topic / Channel | Producer | Consumer(s) | Schema | Idempotent |
|------------|----------------|----------|-------------|--------|------------|
| `user.created` | `users` | `UserService` | `EmailService`, `AnalyticsService` | `UserCreatedEvent` | Yes (by user_id) |
| `order.paid` | `orders` | `PaymentGateway` | `FulfillmentService` | `OrderPaidEvent` | Yes (by order_id) |
| `cart.updated` | in-process | `CartStore` | `CartBadge`, `CartSummary` | `CartState` | N/A |

Extract from actual event publishing and subscription code — do not invent events.

##### 3. Event Schemas

For each event (or event group), document the payload:

```typescript
// user.created
interface UserCreatedEvent {
  event_id: string;        // Unique event identifier (UUID)
  event_type: "user.created";
  timestamp: string;       // ISO 8601
  data: {
    user_id: string;
    email: string;
    registered_via: "email" | "oauth";
  };
  metadata: {
    correlation_id: string;  // Request trace ID
    source: string;          // Producing service name
  };
}
```

Use the project's actual language and types. Note which fields are required vs optional.

##### 4. Topics / Channels / Queues

| Topic / Queue | Partitions | Retention | Consumer Group(s) | Purpose |
|--------------|------------|-----------|-------------------|---------|
| `users` | 3 | 7 days | `email-svc`, `analytics-svc` | User lifecycle events |
| `orders` | 6 | 30 days | `fulfillment-svc` | Order processing |

Include configuration source (env vars, config files, IaC definitions).

##### 5. Producers

For each producer:
- **Location**: file path and function/method
- **Trigger**: what causes the event to be published
- **Serialization**: how the event is serialized before sending
- **Error handling**: what happens if publishing fails (retry, circuit breaker, outbox pattern)
- **Transactional outbox**: if used, describe the pattern (DB transaction + outbox table + poller)

Code example of a typical producer from the project.

##### 6. Consumers

For each consumer:
- **Location**: file path and function/method
- **Concurrency**: how many concurrent consumers (worker count, partition assignment)
- **Processing guarantee**: at-least-once, at-most-once, exactly-once
- **Idempotency strategy**: how duplicate events are handled (idempotency key, deduplication window, upsert)
- **Error handling**: what happens on processing failure
- **Ordering**: does processing depend on event order? How is it guaranteed?

Code example of a typical consumer from the project.

##### 7. Delivery Guarantees

| Guarantee | Strategy | Implementation |
|-----------|----------|----------------|
| At-least-once | Manual ack after processing | Consumer commits offset after handler returns |
| Ordering | Per-key ordering | Partition key = entity ID |
| Idempotency | Dedup by event_id | Check processed_events table before handling |
| Exactly-once | Transactional outbox | DB transaction + outbox poller |

##### 8. Dead Letter Queue (DLQ)

- DLQ topic/queue name and configuration
- When events are sent to DLQ (max retries exceeded, deserialization failure, handler exception)
- Retry policy before DLQ: count, backoff strategy (fixed, exponential, jitter)
- How DLQ events are monitored and reprocessed
- Manual replay mechanism (if available)

If no DLQ exists, note it explicitly and flag as a risk.

##### 9. Event Versioning & Evolution

- How event schemas evolve (backward-compatible additions, schema registry, version field)
- Breaking change process: how old consumers handle new event versions
- Schema validation: compile-time (Protobuf, Avro) or runtime (JSON Schema)
- If no versioning strategy exists, note it and flag as a risk.

##### 10. Monitoring & Observability

- Consumer lag monitoring (Kafka consumer group lag, queue depth)
- Event throughput metrics
- Failed event alerting
- Correlation ID / trace propagation through events
- Dashboard links (if applicable)

##### 11. Testing Events

- How to test producers (unit test: event published with correct payload)
- How to test consumers (unit test: handler processes event correctly)
- Integration testing: in-memory broker, test containers, embedded broker
- How to publish test events manually (CLI tool, admin endpoint, script)

Code example of an event test from the project.

## General Rules

- Language: English
- All events, topics, and schemas must come from actual source code — do not invent
- If the project uses multiple event systems (e.g., Kafka for inter-service + in-process for domain events), document all and explain boundaries
- For frontend-only event systems (CustomEvent, RxJS), adapt the template: skip DLQ, delivery guarantees, and monitoring sections
- If the project has no event-driven patterns, do not generate this file — skip entirely
- After creating, update `.spec/README.md`: add a link under the appropriate section
