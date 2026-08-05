# Active Context

## Current State

- AI context has been moved under `.ai/`, with `AI_CONTEXT.md` as the root entrypoint.
- Architecture documentation has been added based on the current Go source.
- `README.md` has been expanded from a placeholder into onboarding and operation notes.

## Known Risks And Follow-Up Ideas

- There are no automated tests. Good first tests would cover MQTT topic generation and Director API response parsing with `httptest`.
- MQTT command topic parsing in `main.go` uses fixed token indexes without validation. Malformed matching topics could cause incorrect parsing or panics.
- Some Aqua Illumination API JSON unmarshal errors are ignored in `GetAll`.
- `docker-compose.yml` sets `MQTT_HOST=localhost` for the app container, which may not reach the Mosquitto service in normal Docker Compose networking.
- `aidirectortomqtt.exe` is a checked-in build artifact and should not be treated as source.

## Recent Validation

- 2026-08-05: `go test ./...` passed. All packages reported no test files.
- 2026-08-05: `go build ./...` passed.
