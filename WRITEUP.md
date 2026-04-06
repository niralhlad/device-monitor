# WRITEUP

## 1. Time spent and most difficult part

I spent approximately one full Sunday working on this challenge, including normal breaks for meals and short pauses during development.

The most difficult part was clarifying the uptime calculation behavior and validating it against the provided simulator. 
The challenge statement leaves some room for interpretation around how the minute range should be calculated between the first and last heartbeat. 
I initially used a more intuitive inclusive minute-range approach, then adjusted the implementation to align with the simulator behavior for submission correctness.

Another area that required care was keeping the implementation simple while still making the code easy to explain during a walkthrough. 
I chose a small in-memory architecture with clear separation between routing, handlers, services, and registry loading so that the data flow remains easy to follow.

## 2. How I would modify the data model or code to support more metrics

The current design already separates:
- HTTP request handling
- Business logic i.e. Service
- Device registry loading
- In-memory device state

That makes it straightforward to support additional metrics.

At the API level, I would keep the same pattern used in the current implementation:
- Handlers would validate and decode metric-specific requests
- Services would apply business rules and update in-memory aggregates
- The stats read endpoint would compose a response from the aggregated device state

If the number of metrics grew significantly, I would refactor the device state into smaller metric-specific aggregators rather than continuing to grow one large struct. For example, I could separate the logic into a heartbeat aggregator, an upload aggregator, and future metric aggregators behind a consistent interface. That would make the code easier to test, easier to extend, and easier for multiple engineers to work on in parallel.

If longer retention or historical reporting were required, I would also consider moving from a purely in-memory model to a persistent storage layer or time-series oriented design. For the scope of this challenge, the in-memory approach keeps the code simple and easy to reason about, but the current separation of handlers and services already gives a clean path for future extension.

## 3. Runtime complexity

### Startup
On startup, the service loads `devices.csv` and builds an in-memory registry of valid device IDs.

If there are `N` device IDs in the CSV, startup loading is:

- **Time:** `O(N)`
- **Space:** `O(N)`

### Heartbeat ingestion
For each heartbeat request, the service:
- validates the device ID
- normalizes the timestamp to a minute bucket
- checks whether that minute was already seen
- updates aggregate state for the device

This is effectively:

- **Average time per request:** `O(1)`
- **Space growth:** proportional to the number of unique heartbeat minutes stored per device

### Upload stats ingestion
For each upload stats request, the service:
- validates the device ID
- validates the payload
- increments upload count
- adds upload duration to a running total

This is:

- **Time per request:** `O(1)`
- **Space growth:** `O(1)` per device for upload aggregates

### Stats read
For each stats read request, the service reads the already-aggregated in-memory state and computes the current response.

This is:

- **Time per request:** `O(1)` in the current design
- **Space:** no additional meaningful per-request storage

## 4. Design choices

I intentionally kept the implementation in-memory because the challenge emphasizes clarity and correctness over heavy infrastructure. A database is not necessary for the required functionality, and an in-memory approach keeps the code easier to reason about during a walkthrough.

I also separated the code into a few clear layers:
- Registry loading for startup device definitions
- Handlers for HTTP concerns
- Services for business logic
- Versioned route registration

This keeps the request flow understandable:
- Startup loads known device IDs
- Handlers validate and decode input
- Services update or read in-memory device state
- Handlers shape the HTTP response

For logging and operability, I added structured logging, request logging middleware, and panic recovery middleware. This keeps the service easier to debug while still remaining lightweight.

## 5. Trade-Off

A key design trade-off in this solution was choosing simplicity and clarity over heavier abstraction and infrastructure. I used a relatively direct router -> handler -> service flow with in-memory per-device state because it keeps the code small, easy to test, and easy to explain during a walkthrough. This fits the scope of the challenge well, but it also means that moving later to multiple storage backends such as Redis or PostgreSQL would likely require introducing another abstraction layer.

I also chose a single-process in-memory design rather than a persistent or distributed architecture. That keeps the implementation lightweight and appropriate for the challenge, but it means state is lost on restart and the current design is not intended for horizontal scaling across multiple application instances. As a lightweight next step, state could be serialized to a file and reloaded on startup, although for a more robust production design I would prefer a durable shared storage layer.

I also kept the logging middleware intentionally simple. Right now it logs request-level information such as method, path, status code, duration, and remote address in a consistent way for all endpoints. This keeps the middleware lightweight and easy to maintain, but it also means the logs do not yet include richer domain-specific fields such as device_id, metric type, or request correlation details. If this were extended further, I would enhance the middleware and possibly selected handlers to include more structured context for device-related debugging and operational tracing.

Finally, there was a trade-off between a more intuitive real-world uptime definition and the behavior expected by the provided simulator. I aligned the final implementation with the simulator for submission correctness, while recognizing that in a production system I would want the uptime definition to be explicitly clarified and agreed upon to avoid ambiguity.

## 6. AI usage

I used AI assistance during the assignment as a support tool for development, review, and communication. In particular, I used it to help with:
- Structuring the project incrementally
- Discussing design trade-offs and architecture options
- Drafting and refining unit tests
- Improving documentation and code comments
- Thinking through edge cases, especially around uptime behavior and simulator compatibility

I did not use AI as a substitute for implementation ownership. I reviewed, adapted, and validated the suggested approaches while building the solution, and I made the final decisions around code structure, API behavior, testing scope, and simulator alignment. It was most helpful as a way to move faster through iteration and debugging.

## 7. Testing, Security, and deployment notes

### Testing
I Added unit tests for:
- Config loading and validation
- Registry loading
- Handlers
- Services
- Routing
- Middleware

The tests focus on request validation, in-memory state updates, uptime behavior, upload aggregation, and route registration.

### Security
For this challenge, the API is intentionally simple and unauthenticated. In a production setting, I would consider:
- Authentication and authorization
- Rate limiting
- Request size limits
- TLS termination
- Structured audit logging
- More careful input validation and error handling

### Deployment
For a larger deployment, I would consider:
- Containerization
- Health probes
- Metrics collection
- Externalized configuration
- Persistent storage if historical data retention were required

## 8. Alpha prototype structure

The diagram below illustrates how I would structure the alpha prototype. It closely reflects the current implementation, with clear separation between routing, handlers, services, and registry.

This keeps the solution simple, testable, and easy to extend while remaining appropriate for the scope of the challenge.

```text
    Devices / Device Simulator
                    |
                    |  HTTP :6733
                    V
  +-----------------------------------------+
  |         cmd/server / internal/app       |
  |                                         |
  |  1. registry.LoadFromCSV -> registry    |
  |  2. services.New(registry) -> service   |
  |  3. handlers.New(service) -> handler    |
  |  4. start HTTP                          |
  +-----------------------------------------+
                    |
                    V
  +-----------------------------------------+
  |         internal/http                   |
  |   router · middleware (logging/recovery)|
  +-----------------------------------------+
                    |
                    V
  +-----------------------------------------+
  |         internal/handlers               |
  |  POST /heartbeat  POST /stats           |
  |  GET  /stats      GET  /health          |
  +-----------------------------------------+
                    |
                    V
  +-----------------------------------------+
  |         internal/services               |
  |  RecordHeartbeat · RecordUploadStat     |
  |  GetStats · uptime · avg upload         |
  |                                         |
  |  in-memory state per device             |
  |  devices buckets · upload durations     |
  +-----------------------------------------+
                    |
                    V
  +-----------------------------------------+
  |         internal/registry               |
  |  loaded from devices.csv at startup     |
  |  validates device ID per request        |
  +-----------------------------------------+
```