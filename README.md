# hasstool

A fast, standalone CLI for [Home Assistant](https://www.home-assistant.io/). Supports one-shot REST queries and live WebSocket event streaming.

## Installation

### Build from source

Requires [Go 1.21+](https://go.dev/dl/).

```bash
git clone https://github.com/your-username/hasstool
cd hasstool
go build -o hasstool .
```

Move the binary somewhere on your `$PATH`:

```bash
mv hasstool /usr/local/bin/
```

## Configuration

`hasstool` needs your Home Assistant URL and a [long-lived access token](https://developers.home-assistant.io/docs/auth_api/#long-lived-access-token).

Set these as environment variables (recommended):

```bash
export HA_URL=http://homeassistant.local:8123
export HA_TOKEN=your_long_lived_token
```

Or pass them as flags on any command:

```bash
hasstool --url http://192.168.1.10:8123 --token <token> states
```

## Usage

### `states` — query entity state(s)

```bash
# All entities
hasstool states

# Single entity
hasstool states light.living_room

# Filter by domain
hasstool states --domain light
hasstool states --domain switch

# Show a specific attribute instead of state
hasstool states light.living_room --attr brightness
hasstool states --domain light --attr brightness   # skips entities without the attribute

# JSON output
hasstool states --json
hasstool states light.living_room --json
```

### `call` — call a service

```bash
# Basic service call
hasstool call switch.turn_on switch.fan_heater

# Multiple targets
hasstool call light.turn_on light.kitchen light.hallway

# With extra service data
hasstool call light.turn_on light.living_room --data '{"brightness": 128, "color_temp": 300}'

# No target (e.g. scene activation)
hasstool call scene.turn_on --data '{"entity_id": "scene.evening"}'
```

### `watch` — stream state changes via WebSocket

```bash
# All entities
hasstool watch

# Single entity
hasstool watch light.living_room

# JSON output (one object per line)
hasstool watch --json
hasstool watch binary_sensor.front_door --json
```

Press `Ctrl-C` to stop watching.

## License

MIT
