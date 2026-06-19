# xoxa-gateway

<img src="./.github/assets/banner.png" width="600px">

A **Go-powered messaging gateway** for multi-channel delivery (**WhatsApp, SMS**).  
Supports **REST + gRPC APIs**, **pluggable transports**, **NATS** for async processing, **PostgreSQL** for persistence, and **OpenTelemetry** for tracing.

---

## Features

- **Fast APIs** → Gin (REST) + gRPC, sharing one service layer
- **Pluggable transports** → Twilio (SMS) and WhatsApp Cloud API today, register more via the `Transport` interface
- **Async queue** → NATS, with exponential backoff retries and a dead-letter subject
- **Persistence** → PostgreSQL + GORM, with embedded SQL migrations applied automatically at startup
- **Tracing & metrics** → OpenTelemetry (OTLP/gRPC) tracing + Prometheus metrics on both `api` and `worker`
- **Auth** → JWT (HS256 via shared secret, or RS256 via a public key file)
- **Docs** → Swagger UI (`/swagger/index.html`) serving `docs/openapi.yaml`, plus `.proto` for gRPC
- **Tests** → Unit tests (in-memory SQLite + `httptest` mock providers); verified end-to-end against real Postgres/NATS via docker-compose
- **Containerized** → Docker + docker-compose for local dev

---

## Project Layout

```
cmd/
  api/main.go          # HTTP + gRPC server
  worker/main.go       # NATS consumer + retry/DLQ + metrics server
internal/
  http/                # REST handlers, JWT middleware
  grpc/                # gRPC server (generated stubs in proto/)
  service/             # business logic: send, status lookup, retry/backoff
  transport/           # pluggable providers (Twilio, WhatsApp Cloud)
  repo/                # GORM models + migration runner
  config/              # env/.env config loader
  telemetry/           # OTel tracing + Prometheus metrics
pkg/
  dto/                 # request/response structs
migrations/            # SQL migrations (embedded at build time)
proto/                 # gRPC service definitions + generated Go code
docs/                  # OpenAPI spec
.github/workflows/     # CI/CD
Dockerfile             # multi-stage build (api or worker via CMD_PATH build arg)
docker-compose.yml     # Postgres + NATS + api + worker
Makefile               # build/test/run/proto shortcuts
```

---

## Prerequisites

- **Go** 1.25+
- **Docker** & **docker-compose**
- **Make** (optional but recommended)

---

## Setup & Run

### 1. Clone & start dependencies

```bash
git clone https://github.com/<you>/xoxa-gateway.git
cd xoxa-gateway
docker compose up -d postgres nats   # or: make up
```

### 2. Configure env

Copy `.env.example` to `.env` and fill in real credentials:

```bash
cp .env.example .env
```

```env
PORT=8080
GRPC_PORT=9090
DATABASE_URL=postgres://xoxa:xoxa@localhost:5432/xoxa?sslmode=disable
NATS_URL=nats://localhost:4222
JWT_ISSUER=xoxa-gateway
JWT_AUDIENCE=internal
JWT_SECRET=supersecret           # or set JWT_PUBLIC_KEY_PATH for RS256
TWILIO_ACCOUNT_SID=
TWILIO_AUTH_TOKEN=
TWILIO_FROM_NUMBER=
WHATSAPP_TOKEN=
WHATSAPP_PHONE_NUMBER_ID=
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

### 3. Run API

```bash
go run ./cmd/api
```

REST: `http://localhost:8080`  
gRPC: `localhost:9090`  
Swagger: `/swagger/index.html`

### 4. Run Worker

```bash
go run ./cmd/worker
```

Worker metrics/health: `http://localhost:9091/metrics`, `/healthz`

### Or run everything in Docker

```bash
docker compose up -d --build
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

- **Metrics** → Prometheus at `/metrics` on the API (`:8080`) and worker (`:9091`)
- **Tracing** → OTLP/gRPC export to Jaeger/Tempo when `OTEL_EXPORTER_OTLP_ENDPOINT` is set (no-op otherwise)
- **Logs** → Gin access logs + structured worker logs to stdout

---

## Testing

```bash
go test ./... -v
```

Transport providers are tested against `httptest` mock servers; the service and HTTP layers use an in-memory SQLite database, so no external services are required. The full stack (Postgres + NATS + real provider HTTP calls) has been validated manually via `docker compose up -d --build`.

---

## Roadmap

- [ ] Telegram transport
- [ ] Delivery callback ingestion (webhooks)
- [ ] Rate-limiting per transport
- [ ] Multi-tenant namespaces
- [ ] Dead-letter queue viewer / replay tooling

---

## License

MIT
