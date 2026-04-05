# device-monitor
A small Go HTTP service that loads device IDs from `devices.csv`, accepts heartbeat and upload telemetry, and returns per-device uptime and average upload duration.

## High-level design

- `cmd/server/main.go` bootstraps the application and handles graceful shutdown.
- `internal/config` wires config, repository, services, handlers, and routes.
- `internal/handlers` owns HTTP decoding, validation, and response shaping.
- `internal/services` owns business rules.
- `internal/registry` owns in-memory persistence and CSV loading.
- `internal/http` registers versioned routes for `v1` and other version when required.

## Supported Endpoint

- `GET /health` : Get the health status of the server.

## Local run

Start the API Server:

```bash
go run ./cmd/server
```

Run the full suite:
```bash
go run test ./...
```