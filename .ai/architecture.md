# Architecture

`aidirectortomqtt` is a polling and command bridge for Aqua Illumination Director lighting controllers.

At runtime it:

1. Loads configuration from environment/settings.
2. Connects to an MQTT broker.
3. Reads current state from the Aqua Illumination Director HTTP API.
4. Publishes MQTT state topics under `aimqtt/`.
5. Publishes Home Assistant MQTT discovery configs under `homeassistant/`.
6. Subscribes to command topics and writes changes back to the Director HTTP API.
7. Polls Director state every second and republishes changed values.
8. Refreshes discovery/configuration every minute.

## Package Boundaries

### `main.go`

`main.go` is the composition root. It owns process startup, logging, configuration loading, MQTT client options, topic subscriptions, initial state publication, polling loops, and signal handling.

Command subscriptions parse MQTT topic tokens to identify the target group, device, and color. Each command calls an Aqua Illumination write function, then immediately reloads Director state and republishes MQTT/Home Assistant data.

### `AquaIllumination/`

This package owns the Director HTTP API boundary.

- `director.go` defines API response models and the aggregated `Director` view used by the rest of the app.
- `update.go` reads Director state from endpoints such as `controller/version`, `tanks`, `devices/statistics`, `devices/{id}/info`, `groups/{id}/led/intensity/colors`, and `groups/{id}/mode`.
- `commands.go` sends `PUT` requests for manual mode and LED intensity changes.

The important architectural detail is that `GetAll` builds a complete in-memory snapshot from several HTTP calls. Other packages should depend on the `Director` aggregate instead of making ad hoc Director API calls.

### `aimqtt/`

This package owns MQTT publishing contracts.

- `UpdateMQTT` publishes runtime state under `aimqtt/{controller}/tank/{tank}/group/{group}/device/{device}/...`.
- `UpdateHomeAssistant` publishes MQTT discovery config under `homeassistant/{platform}/{device}/{entity}/config`.
- `AiMqtt.data` caches the last payload for each full topic so unchanged values are not republished unless `forceUpdate` is true.

The package also maps AI color channel names to RGB values for Home Assistant light entities.

### `settings/`

This package maps settings keys into a small `Config` struct:

- `MQTT_HOST`
- `MQTT_PORT`
- `MQTT_USER`
- `MQTT_PASSWORD`
- `DIRECTOR_HOST`

`ConfigFile` is read in `main.go` and passed to `github.com/cjburchell/settings-go`.

## Data Flow

```text
Aqua Illumination Director HTTP API
        |
        v
AquaIllumination.GetAll
        |
        v
AquaIllumination.Director snapshot
        |
        +--> aimqtt.UpdateMQTT ----------> MQTT state topics
        |
        +--> aimqtt.UpdateHomeAssistant -> Home Assistant discovery topics

Home Assistant / MQTT command topics
        |
        v
main.go subscriptions
        |
        v
AquaIllumination.SetManualMode / ToggleLight / SetIntensity
        |
        v
Director HTTP PUT endpoints
```

## MQTT Contract

State topics are rooted at `aimqtt/`. Discovery topics are rooted at `homeassistant/`.

Command topics currently handled:

- `aimqtt/+/tank/+/group/+/device/+/manualModeCommand`
- `aimqtt/+/tank/+/group/+/device/+/+/toggle`
- `aimqtt/+/tank/+/group/+/device/+/+/setintensity`

Topic token indexes are parsed directly in `main.go`; changes to topic shape require coordinated updates to subscriptions, parsing, state publishing, and Home Assistant discovery payloads.

## Operational Notes

- MQTT reconnect is enabled with short timeouts and a fixed client ID of `ai-mqtt`.
- The bridge publishes `aimqtt/status=online` during normal updates and `offline` when the config refresh loop receives an interrupt.
- There are no tests currently in the repository.
