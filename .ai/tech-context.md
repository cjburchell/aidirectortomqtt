# Tech Context

## Stack

- Language: Go 1.20
- Module: `gitlab.com/cjburchell/aidirectortomqtt`
- MQTT client: `github.com/eclipse/paho.mqtt.golang`
- Settings: `github.com/cjburchell/settings-go`
- Environment helper: `github.com/cjburchell/tools-go/env`
- Logging: `github.com/cjburchell/uatu-go`
- Error wrapping: `github.com/pkg/errors`

## Build And Validation

Run from the repository root:

```powershell
go test ./...
go build ./...
docker build .
```

GitLab CI runs:

- `hadolint Dockerfile`
- Docker image builds and registry pushes for selected branches/tags.

## Runtime Configuration

The app reads settings through `settings-go`; `main.go` passes `ConfigFile` from the environment to the settings loader.

Supported app settings:

| Setting | Default | Purpose |
| --- | --- | --- |
| `MQTT_HOST` | `localhost` | MQTT broker host |
| `MQTT_PORT` | `1883` | MQTT broker port |
| `MQTT_USER` | empty | MQTT username |
| `MQTT_PASSWORD` | empty | MQTT password |
| `DIRECTOR_HOST` | empty | Aqua Illumination Director host/IP |
| `MinLogLevel` | `INFO` | Logger minimum level |
| `ConfigFile` | empty | Optional settings file path consumed by `settings-go` |

## Docker

`Dockerfile` uses a Go builder image and copies a static Linux binary into a `scratch` runtime image.

`docker-compose.yml` defines:

- Home Assistant
- Eclipse Mosquitto
- `ai-mqtt`, built from this repository

Review networking before relying on the current compose defaults. The `ai-mqtt` service sets `MQTT_HOST=localhost`, which means localhost inside the app container unless Docker networking is changed.

## Repository Notes

- `aidirectortomqtt.exe` is a committed binary artifact.
- `.idea/` is ignored.
- No test files are present as of this context update.
- Existing package names include `AquaIllumination` with capital letters; keep imports consistent unless doing a deliberate package rename.
