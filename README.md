# aidirectortomqtt

`aidirectortomqtt` bridges an Aqua Illumination Director lighting controller to MQTT and Home Assistant.

The service polls the Director HTTP API, publishes device state to MQTT topics, publishes Home Assistant MQTT discovery configuration, and listens for MQTT commands that toggle manual mode or change LED channel intensity.

## What It Does

- Reads tanks, groups, devices, device statistics, group LED colors, and group mode from the Aqua Illumination Director API.
- Publishes state under `aimqtt/` MQTT topics.
- Publishes Home Assistant discovery payloads under `homeassistant/`.
- Handles Home Assistant/MQTT commands for:
  - manual mode
  - light channel on/off
  - light channel brightness/intensity
- Polls state every second and refreshes discovery/configuration every minute.

## Configuration

Configuration is loaded through `github.com/cjburchell/settings-go`. `ConfigFile` may be set to point at a settings file; otherwise environment/default values are used.

| Setting | Default | Description |
| --- | --- | --- |
| `MQTT_HOST` | `localhost` | MQTT broker host |
| `MQTT_PORT` | `1883` | MQTT broker port |
| `MQTT_USER` | empty | MQTT username |
| `MQTT_PASSWORD` | empty | MQTT password |
| `DIRECTOR_HOST` | empty | Aqua Illumination Director host/IP |
| `MinLogLevel` | `INFO` | Minimum log level |
| `ConfigFile` | empty | Optional settings file path |

## Build

```powershell
go build ./...
```

To build the container image:

```powershell
docker build .
```

## Run

Run locally with environment variables:

```powershell
$env:MQTT_HOST="localhost"
$env:MQTT_PORT="1883"
$env:DIRECTOR_HOST="192.168.3.104"
go run .
```

Or use Docker Compose as a starting point:

```powershell
docker compose up --build
```

Review `docker-compose.yml` before relying on it unchanged. In normal Docker Compose networking, `MQTT_HOST=localhost` inside the app container refers to that container, not the Mosquitto service.

## MQTT Topics

State topics are rooted at `aimqtt/` and include controller, tank, group, device, and channel identifiers.

The app subscribes to these command topic patterns:

```text
aimqtt/+/tank/+/group/+/device/+/manualModeCommand
aimqtt/+/tank/+/group/+/device/+/+/toggle
aimqtt/+/tank/+/group/+/device/+/+/setintensity
```

Home Assistant discovery topics are published under:

```text
homeassistant/{platform}/{device}/{entity}/config
```

## Project Layout

```text
main.go              Application wiring, MQTT subscriptions, polling loops
AquaIllumination/    Director API models, reads, and commands
aimqtt/              MQTT state publishing and Home Assistant discovery
settings/            Runtime configuration mapping
.ai/                 AI development context and architecture notes
```

## Validation

```powershell
go test ./...
go build ./...
```

There are currently no automated tests in the repository.

## AI Development Context

AI-oriented repository context starts at [AI_CONTEXT.md](AI_CONTEXT.md). Supporting memory and architecture notes live in [.ai/](.ai/).
