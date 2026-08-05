# AI Context

This is a Go bridge between an Aqua Illumination Director controller, MQTT, and Home Assistant MQTT discovery.

Future AI sessions should read this file first, then load only the supporting context needed for the task.

## Read First

- [README.md](README.md) for human-facing setup, behavior, and runtime configuration.
- [.ai/architecture.md](.ai/architecture.md) before changing control flow, MQTT topics, Home Assistant discovery, or Aqua Illumination API access.
- [.ai/tech-context.md](.ai/tech-context.md) before changing build, dependency, Docker, CI, or validation behavior.
- [.ai/active-context.md](.ai/active-context.md) for current notes, risks, and likely follow-up work.

## Source Layout

- `main.go` wires configuration, logging, MQTT subscriptions, startup discovery, polling, and shutdown handling.
- `AquaIllumination/` contains HTTP API models and client functions for reading and commanding the AI Director.
- `aimqtt/` converts AI Director state into MQTT state topics and Home Assistant discovery payloads.
- `settings/` maps runtime settings to the application config struct.

## Do Not Edit Casually

- `aidirectortomqtt.exe` is a checked-in build artifact.
- `.git/`, `.idea/`, and generated CI artifacts should not be edited by AI sessions.
- `go.sum` should only change as a consequence of dependency changes through Go tooling.

## Validation

Use the smallest relevant validation for the change:

```powershell
go test ./...
go build ./...
docker build .
```

Docker linting is performed in GitLab CI with `hadolint Dockerfile`.

## Working Expectations

- Inspect the relevant package before editing.
- Preserve MQTT topic structure unless the task explicitly changes integration contracts.
- Preserve Home Assistant discovery JSON fields and unique IDs carefully; changing them can create duplicate or orphaned entities.
- Update `README.md` for user-facing behavior and `.ai/architecture.md` for architectural changes.
