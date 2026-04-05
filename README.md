# device-monitor
A small Go HTTP service that loads device IDs from `devices.csv`, accepts heartbeat and upload telemetry, and returns per-device uptime and average upload duration.

## High-level design

- `cmd/server/main.go` bootstraps the application and handles graceful shutdown.
- `internal/app` wires config, logger, registry, handlers, and routes.
- `internal/config` loads and validates environment-based configuration.
- `internal/handlers` owns HTTP decoding, validation, and response shaping.
- `internal/services` owns business rules and in-memory device state.
- `internal/registry` loads valid device IDs from CSV at startup.
- `internal/http` registers versioned routes for `v1`.

## Supported Endpoints

- `GET /health` : Get the health status of the server.
- `POST /api/v1/devices/{device_id}/heartbeat` : Register a heartbeat for a known device.
- `GET /api/v1/devices/{device_id}/stats` : Return current device uptime statistics.

## Local run

Start the API Server:

```bash
go run ./cmd/server
```

Run the full suite:
```bash
go run test ./...
```