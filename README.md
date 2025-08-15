# xoxa-gateway

<img src="./.github/assets/banner.png" width="600px">

A **Go-powered messaging gateway** for multi-channel delivery (**WhatsApp, SMS, Telegram**).  
Supports **REST + gRPC APIs**, **pluggable transports**, **NATS** for async processing, **PostgreSQL** for persistence, and **OpenTelemetry** for tracing.

---

## Features

- **Fast APIs** → Gin (REST) + gRPC
- **Pluggable transports** → Twilio, WhatsApp Cloud, etc.
- **Async queue** → NATS (streaming, retries, DLQ)
- **Persistence** → PostgreSQL + GORM
- **Tracing & metrics** → OpenTelemetry + Prometheus
- **Auth** → JWT (HS256 or RSA)
- **Docs** → Swagger/OpenAPI (REST) + `.proto` (gRPC)
- **Tests** → Unit + Integration (Testcontainers)
- **Containerized** → Docker + docker-compose for local dev

---

## Project Layout

```
cmd/
  api/main.go          # HTTP server
  worker/main.go       # NATS consumers
internal/
  http/                # REST handlers
  grpc/                # gRPC server
  service/             # business logic
  transport/           # pluggable providers
  repo/                # database access
  config/              # env/config loader
  telemetry/           # tracing & metrics
pkg/
  dto/                 # request/response structs
migrations/            # SQL migrations
proto/                 # gRPC service definitions
docs/                  # API docs
.github/workflows/     # CI/CD
```

---

## Prerequisites

- **Go** 1.22+
- **Docker** & **docker-compose**
- **Make** (optional but recommended)

---

## Setup & Run

### 1. Clone & start services

```bash
git clone https://github.com/<you>/xoxa-gateway.git
cd xoxa-gateway
docker compose up -d  # starts Postgres + NATS
```

### 2. Configure env

Create `.env`:

```env
PORT=8080
DATABASE_URL=postgres://xoxa:xoxa@localhost:5432/xoxa?sslmode=disable
NATS_URL=nats://localhost:4222
JWT_ISSUER=xoxa-gateway
JWT_AUDIENCE=internal
JWT_SECRET=supersecret   # or set RSA public/private keys
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

### 3. Run API

```bash
go run ./cmd/api
```

REST: `http://localhost:8080`  
Swagger: `/swagger/index.html`

### 4. Run Worker

```bash
go run ./cmd/worker
```

---

## API Quick Reference

**Health check**

```bash
curl http://localhost:8080/healthz
# { "ok": true }
```

**Send a message**

```bash
curl -X POST http://localhost:8080/api/v1/messages   -H "Authorization: Bearer <token>"   -H "Content-Type: application/json"   -d '{
        "channel": "whatsapp",
        "to": "+27000000000",
        "text": "Hello from xoxa-gateway!"
      }'
```

**Response:**

```json
{ "id": "msg_123", "status": "queued" }
```

**Get message status**

```bash
curl http://localhost:8080/api/v1/messages/msg_123
```

---

## Transport Interface (for new providers)

```go
type Transport interface {
    Name() string
    Send(ctx context.Context, msg Message) (providerID string, finalStatus Status, err error)
}
```

Implement this in `internal/transport/<name>.go`, then register it in the transport registry.

---

## Observability

- **Metrics** → Prometheus at `/metrics`
- **Tracing** → OTLP to Jaeger/Tempo
- **Logs** → Structured JSON with request IDs

---

## Testing

```bash
go test ./... -v
```

---

## Roadmap

- [ ] Delivery callback ingestion (webhooks)
- [ ] Rate-limiting per transport
- [ ] Multi-tenant namespaces
- [ ] Dead-letter queue viewer

---

## License

MIT
