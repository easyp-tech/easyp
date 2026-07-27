<!-- scope: **/*worker*, **/*job*, **/*queue*, **/*cron* -->
# Background Jobs Documentation Template

## What This Generates

- `.spec/BACKGROUND_JOBS.md` — documents the background job system, job inventory, retry strategy, and scaling approach

## Instructions

Analyze the project to identify any background job or async task processing. Look for:
- Job frameworks (Sidekiq, Celery, Temporal, Asynq, Bull, Hangfire, Faktory)
- Custom worker implementations (goroutines + channels, threads + queues)
- Cron/scheduler configurations (crontab, systemd timers, Kubernetes CronJobs, `robfig/cron`)
- Message queue consumers (Kafka consumers, RabbitMQ subscribers, NATS workers, SQS listeners)

If no background job system is found, **skip this template entirely** — do not generate an empty document.

### File: .spec/BACKGROUND_JOBS.md

#### Structure:

##### 1. Overview
- **System**: which job framework or custom implementation is used
- **Transport**: underlying queue/broker (Redis, RabbitMQ, Kafka, PostgreSQL, in-memory)
- **Architecture pattern**: pull-based (workers poll queue) or push-based (broker dispatches)
- **Entry point**: where workers are started (separate binary, same process, sidecar)

##### 2. Job Inventory

Table of all registered jobs:

| Job Name | Trigger | Frequency / Event | Timeout | Retry | DLQ | Priority |
|----------|---------|-------------------|---------|-------|-----|----------|
| `SendWelcomeEmail` | Event: user.created | On event | 30s | 3× | ✅ | normal |
| `CleanupExpiredSessions` | Cron | Every hour | 5m | 1× | ❌ | low |
| `ProcessPayment` | Event: order.placed | On event | 60s | 5× exponential | ✅ | high |

Trigger types: `Cron` (scheduled), `Event` (message/webhook), `Manual` (API/CLI), `Cascade` (triggered by another job).

##### 3. Architecture

ASCII diagram showing the job processing flow:

```
Producer → Queue/Broker → Worker Pool → Result Store
                ↓ (on failure)
           Dead Letter Queue → Alert
```

For each component:
- **Producer**: which services enqueue jobs, how (SDK call, publish message, HTTP endpoint)
- **Queue**: queue names, partitioning, priority levels
- **Worker**: concurrency model (process per job, thread pool, goroutine pool), startup configuration
- **Result**: where job results are stored (database, cache, nowhere), how callers check completion

##### 4. Retry & Error Handling
- **Retry policy**: max retries per job, backoff strategy (fixed, exponential, custom)
- **Backoff configuration**: initial delay, multiplier, max delay, jitter
- **Dead letter queue (DLQ)**: which jobs have DLQ, how DLQ is monitored, manual replay procedure
- **Idempotency**: which jobs MUST be idempotent, how idempotency is enforced (unique keys, deduplication)
- **Error classification**: transient errors (retry) vs permanent errors (DLQ), how the system distinguishes them
- **Poison messages**: how malformed/unprocessable messages are handled

##### 5. Concurrency & Ordering
- **Worker count**: how many workers per queue, how it's configured
- **Parallelism**: can the same job type run in parallel? Are there mutex/lock constraints?
- **Ordering guarantees**: FIFO per queue? Per partition key? No guarantees?
- **Rate limiting**: are any jobs rate-limited (e.g., email sending, API calls)?
- **Distributed locking**: if jobs use locks, which lock mechanism (Redis, DB advisory locks, etcd)

##### 6. Monitoring
- **Metrics**: which job metrics are collected
  - Queue depth (pending jobs)
  - Processing duration (p50, p95, p99)
  - Success/failure rate
  - DLQ size
- **Dashboards**: Grafana/Datadog dashboard links or query examples
- **Alerting**: alert rules (e.g., "DLQ > 100", "queue depth > 1000 for 5min", "job duration > 2× p99")
- **Logging**: what is logged per job execution (start, end, duration, error, retry count)

##### 7. Scaling
- **Horizontal scaling**: how to add more workers (replicas, autoscaling, manual)
- **Queue partitioning**: are queues partitioned? By what key?
- **Priority queues**: which priority levels exist, how workers consume them (weighted, strict)
- **Backpressure**: what happens when the queue is full (block producer, drop, overflow queue)
- **Resource limits**: CPU/memory per worker, connection pool sizing

##### 8. Development
Step-by-step guide to add a new background job:
1. Define job payload struct/schema
2. Implement handler function
3. Register handler with the worker
4. Create producer (enqueue function)
5. Add retry/DLQ configuration
6. Write tests (unit + integration)
7. Add monitoring (metrics, alerts)
8. Deploy (worker restart required?)

- **Local testing**: how to run workers locally, how to enqueue test jobs
- **Job simulation**: how to trigger a job manually for debugging (CLI, admin API, console)

## General Rules

- Language: English
- Use actual job names, queue names, and code references from the project — do not invent examples
- If no background job system exists, do not generate this file
- For message queue consumers that are part of the infrastructure layer, defer to `infrastructure.md` for broker-level docs (connection, topology). This template owns the job-level details (handler, retry, monitoring).
