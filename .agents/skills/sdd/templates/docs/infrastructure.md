<!-- scope: **/infra/*, **/redis*, **/kafka*, **/nats* -->
# Infrastructure Documentation Template

## What This Generates

One `.spec/<COMPONENT>.md` file per significant infrastructure component discovered in the project. Examples:
- `.spec/OBSERVABILITY.md` — OpenTelemetry, Prometheus, Grafana, Jaeger, etc.
- `.spec/REDIS.md` — Redis caching, sessions, leader election
- `.spec/TRAEFIK.md` / `.spec/NGINX.md` — reverse proxy
- `.spec/QUEUES.md` / `.spec/KAFKA.md` — message queues
- `.spec/LEADER_ELECTION.md` — distributed coordination
- `.spec/SEARCH.md` — Elasticsearch, Meilisearch, etc.

## Instructions

You are a technical documentarian. Create infrastructure documentation for the project in the `.spec/` directory.
Analyze: `docker-compose.yml`, `configs/`, caching adapters, monitoring setup, proxy config, and other infrastructure components.

### Step 1: Identify Infrastructure Components

Check for the presence of:
- **Observability**: OpenTelemetry, Prometheus, Grafana, Jaeger, Tempo, Loki, Datadog, New Relic, Sentry, Pyroscope
- **Caching**: Redis, Memcached, in-memory cache
- **Message Queues**: Kafka, RabbitMQ, NATS, Redis Pub/Sub
- **Reverse Proxy**: Traefik, Nginx, Caddy, Envoy
- **Search**: Elasticsearch, Meilisearch, Typesense
- **Distributed Coordination**: Leader election, distributed locks, service discovery
- **CDN / Object Storage**: CloudFront, Cloudflare, S3

For each discovered component, create a separate `.spec/COMPONENT_NAME.md` file.

### Step 2: Create Documents

For each infrastructure component, use the following template:

---

#### Section 1: Overview
- One sentence: what it is and why it's used
- ASCII diagram: how the component connects to the rest of the system

#### Section 2: Architecture / Components
If the component has multiple parts (e.g., Grafana stack = Tempo + Loki + Mimir + Alloy):
- For each part: role, configuration file, endpoints (ports)

If the component is standalone (e.g., Redis):
- Usage diagram (which services connect, what data they store)

#### Section 3: Use Cases (for general-purpose components like Redis)
Table:
| Use Case | Key Pattern | TTL | Purpose |
|----------|-------------|-----|---------|

#### Section 4: File Structure
```
path/to/adapter/
├── file1.go    # Description
├── file2.go    # Description
└── ...
```

#### Section 5: Interfaces
Interface definitions from the code (contracts.go, interfaces, etc.).
Show the full interface code.

#### Section 6: Implementation Details

For each use case / method:
- Name
- Usage example (code)
- Configuration
- Key parameters (TTL, retry policy, etc.)

**Type-specific sections:**

**For Observability (OBSERVABILITY.md):**
- Backend Integration: how tracing/metrics are initialized in code
- Instrumentation Layers: table (Layer | Package | Technology | Span Names)
- Trace-Log Correlation: how trace_id gets into logs
- Span Hierarchy Example: span tree for a typical request
- Grafana / Dashboard access and key queries

**For Cache/Redis (REDIS.md):**
- Generic Cache API (Set/Get/Del)
- Session Storage: keys, TTL, format
- Cache-Aside / Read-Through pattern
- Invalidation strategy
- Algorithm (for leader election and similar): steps + Lua scripts

**For Reverse Proxy (TRAEFIK.md / NGINX.md):**
- Service URLs table
- How to customize (prefixes, domains)
- Configuration (labels, config files)
- Ports exposed to host
- Troubleshooting

**For Message Queues (QUEUES.md / KAFKA.md):**
- Topics / Exchanges / Subjects
- Producer / Consumer pattern
- Retry / DLQ policy
- Serialization format

**For Leader Election (LEADER_ELECTION.md):**
- Algorithm (SETNX, Consul, etcd)
- Failover scenarios
- Heartbeat mechanism
- Parameters table (TTL, heartbeat interval, namespace)

#### Section 7: Configuration

Docker Compose fragment:
```yaml
service_name:
  image: ...
  ports: ...
  environment: ...
  volumes: ...
```

Application config (config.yml / .env):
```yaml
component:
  url: "..."
  option: "..."
```

#### Section 8: Testing
- How the component is tested (integration tests, test containers)
- Test example

#### Section 9: Observability (for non-tracing components)
- How the component is traced (redisotel, otelhttp, etc.)
- Health check

#### Section 10: Verification Commands
```bash
# Check connection
docker exec container_name ...
# View data
...
# Monitoring
...
```

#### Section 11: Key Files
Table:
| File | Description |
|------|-------------|

## General Rules

- Create a separate file for each significant infrastructure component
- If a component is trivial (1 config, no code) — do not create a separate file, mention it in `ARCHITECTURE.md`
- All code examples and configs — from the actual project
- Docker Compose fragments — from the actual `docker-compose.yml`
- After creating files, update `.spec/README.md`: add links under the Infrastructure section
- If no infrastructure components are found, do not generate any files — skip this template entirely
