# device-monitor

A small Go HTTP service that loads device IDs from `data/devices.csv`, accepts device heartbeat and upload telemetry, and returns per-device uptime and average upload duration statistics.

## High-level design

This project follows a **layered architecture** with clear separation of responsibilities across the application lifecycle, HTTP layer, business logic layer, and infrastructure-style support modules. The structure is designed to keep the codebase easy to read, test, and extend.

## Project structure

```text
device-monitor
├── cmd/
│   └── server/
│       └── main.go                 # application entrypoint
├── data/
│   └── devices.csv                 # valid device IDs loaded at startup
├── internal/
│   ├── app/                        # dependency wiring and bootstrap
│   ├── config/                     # environment-driven configuration
│   ├── handlers/                   # HTTP request validation and responses
│   ├── http/                       # router and middleware
│   ├── registry/                   # device CSV loading and lookup
│   └── services/                   # in-memory aggregation logic
├── .env.example                    # env example
├── Makefile                        # common development commands
├── README.md
├── go.mod
├── results.txt                     # simulator output from a successful run
```

## API summary

### Health check

**GET** `/health`

Returns service health information.

Response:

```bash
{
    "status": "ok",
    "service": "device-monitor",
    "environment": "dev"
}
```

### Record heartbeat

**POST** `/api/v1/devices/{device_id}/heartbeat`

Records a heartbeat timestamp for a known device.

Request body:

```json
{
  "sent_at": "2026-04-05T12:00:10Z"
}
```

Success response:

- `204 No Content`

### Record upload stats

**POST** `/api/v1/devices/{device_id}/stats`

Records upload telemetry for a known device.

Request body:

```json
{
  "sent_at": "2026-04-05T12:00:10Z",
  "upload_time": 60000000000
}
```

Notes:

- `upload_time` is stored as a duration in nanoseconds
- `sent_at` is validated by the handler, although the current aggregation logic only uses `upload_time`

Success response:

- `204 No Content`

### Get device stats

**GET** `/api/v1/devices/{device_id}/stats`

Returns the currently aggregated statistics for a known device.

Example response:

```json
{
  "uptime": 99.79167,
  "avg_upload_time": "3m7.893379134s"
}
```

## Current behavior

- Device IDs are loaded from `data/devices.csv` during startup.
- Unknown devices are rejected with `404 Not Found`.
- Heartbeats are stored in memory per device.
- Upload durations are stored in memory per device.
- Uptime is calculated from heartbeat minute buckets.
- Average upload time is calculated from the recorded upload durations.
- The service keeps runtime state in memory for simplicity.

## Assumptions

- Heartbeat telemetry and upload telemetry are treated as separate signals.
- The heartbeat endpoint drives uptime calculation.
- The upload stats endpoint drives average upload time calculation.
- Device definitions are loaded once at startup from `data/devices.csv`.
- The implementation is intentionally simple and in-memory for clarity and ease of explanation.
- Uptime is based on heartbeat events, consistent with the provided device simulator.

## Makefile commands

The project includes a `Makefile` for common development tasks.

### Available commands

```bash
make help           # Open Makefile help
make run            # Run the server
make test           # Run all tests
make test-no-cache  # Run all tests without cache
make fmt            # Format Go code
make vet            # Run go vet
make clean          # Remove build artifacts
make build          # Build the server binary
make simulator      # Run the device simulator
```

## Running the project

Before running the project, copy the example environment file and update it with values appropriate for your local setup:

```bash
cp .env.example .env
```

Set the following variables in .env:
```bash
SERVICE_NAME=YOUR_SERVICE_NAME
ENVIRONMENT={development, staging, production}
HTTP_PORT=PORT_NUMBER
LOG_LEVEL={debug, info, warning, error}
LOG_FORMAT={text or json}
DEVICES_CSV_PATH=PATH_TO_DEVICES.CSV
```

Field descriptions:
- `SERVICE_NAME`: name of the service used in logs and application metadata
- `ENVIRONMENT`: application environment such as `development`, `staging`, or `production`
- `HTTP_PORT`: port on which the HTTP server will listen
- `LOG_LEVEL`: log verbosity level such as `debug`, `info`, `warning`, or `error`
- `LOG_FORMAT`: log output format, usually `text` or `json`
- `DEVICES_CSV_PATH`: path to the `devices.csv` file loaded at startup

### 1. Install dependencies

```bash
go mod download
```

### 2. Start the server

Using Go directly:

```bash
go run ./cmd/server
```

Or with Make:

```bash
make run
```

The service starts on:

```text
http://localhost:6733
```

### 3. Run tests

```bash
go test ./...
```

Or:

```bash
make test
```

### 4. Build the binary

```bash
make build
```

### 5. Running the simulator

Before running the simulator, download the simulator binary and place it in the project root directory.

If you plan to use `make simulator`, ensure the simulator filename in the `Makefile` matches the binary on your machine. For example, if the binary is named `device-simulator-mac-arm64`, the `Makefile` should reference that exact filename.

Open a separate terminal and run:

```bash
make simulator
```

The simulator will send device data to the service, query the stats endpoint for each device, and generate a `results.txt` file in the project root.

This repository includes a `results.txt` file from a successful simulator run for reference.

## Example curl commands

Heartbeat:

```bash
curl -i -X POST http://localhost:6733/api/v1/devices/60-6b-44-84-dc-64/heartbeat \
  -H 'Content-Type: application/json' \
  -d '{"sent_at":"2026-04-05T12:00:10Z"}'
```

Upload stats:

```bash
curl -i -X POST http://localhost:6733/api/v1/devices/60-6b-44-84-dc-64/stats \
  -H 'Content-Type: application/json' \
  -d '{"sent_at":"2026-04-05T12:00:10Z","upload_time":60000000000}'
```

Read stats:

```bash
curl -s http://localhost:6733/api/v1/devices/60-6b-44-84-dc-64/stats
```