package ha

import "encoding/json"

// WSMessage is the envelope for all WebSocket messages.
type WSMessage struct {
	ID   int             `json:"id,omitempty"`
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"`
}

// AuthMessage is sent to authenticate the WebSocket connection.
type AuthMessage struct {
	Type        string `json:"type"`
	AccessToken string `json:"access_token"`
}

// SubscribeMessage subscribes to an event type.
type SubscribeMessage struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	EventType string `json:"event_type,omitempty"`
}

// CallServiceMessage calls a HA service.
type CallServiceMessage struct {
	ID          int            `json:"id"`
	Type        string         `json:"type"`
	Domain      string         `json:"domain"`
	Service     string         `json:"service"`
	ServiceData map[string]any `json:"service_data,omitempty"`
	Target      *Target        `json:"target,omitempty"`
}

// Target specifies which entities a service call targets.
type Target struct {
	EntityID []string `json:"entity_id,omitempty"`
}

// ResultMessage is the response to a command.
type ResultMessage struct {
	ID      int             `json:"id"`
	Type    string          `json:"type"`
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResultError    `json:"error,omitempty"`
}

// ResultError contains error details from a failed command.
type ResultError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// EventMessage wraps a state_changed event.
type EventMessage struct {
	ID    int   `json:"id"`
	Type  string `json:"type"`
	Event Event `json:"event"`
}

// Event is the inner event payload.
type Event struct {
	EventType string          `json:"event_type"`
	Data      json.RawMessage `json:"data"`
}

// StateChangedData is the data payload for state_changed events.
type StateChangedData struct {
	EntityID string  `json:"entity_id"`
	OldState *State  `json:"old_state"`
	NewState *State  `json:"new_state"`
}

// State represents a HA entity state.
type State struct {
	EntityID    string            `json:"entity_id"`
	State       string            `json:"state"`
	Attributes  map[string]any    `json:"attributes"`
	LastChanged string            `json:"last_changed"`
	LastUpdated string            `json:"last_updated"`
}
