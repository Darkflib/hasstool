package ha

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// Client holds connection details for a Home Assistant instance.
type Client struct {
	BaseURL string
	Token   string
	msgID   atomic.Int32
}

// NewClient creates a new HA client. baseURL should be like "http://homeassistant.local:8123".
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
	}
}

func (c *Client) nextID() int {
	return int(c.msgID.Add(1))
}

// httpHeader returns the Authorization header for REST requests.
func (c *Client) httpHeader() http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+c.Token)
	h.Set("Content-Type", "application/json")
	return h
}

// GetStates fetches all entity states via the REST API.
func (c *Client) GetStates(ctx context.Context) ([]State, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/states", nil)
	if err != nil {
		return nil, err
	}
	for k, v := range c.httpHeader() {
		req.Header[k] = v
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET /api/states: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/states: unexpected status %s", resp.Status)
	}

	var states []State
	if err := json.NewDecoder(resp.Body).Decode(&states); err != nil {
		return nil, fmt.Errorf("decode states: %w", err)
	}
	return states, nil
}

// GetState fetches a single entity state via the REST API.
func (c *Client) GetState(ctx context.Context, entityID string) (*State, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/states/"+entityID, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range c.httpHeader() {
		req.Header[k] = v
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET /api/states/%s: %w", entityID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("entity %q not found", entityID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/states/%s: unexpected status %s", entityID, resp.Status)
	}

	var state State
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("decode state: %w", err)
	}
	return &state, nil
}

// wsURL converts the base HTTP URL to a WebSocket URL.
func (c *Client) wsURL() string {
	u := strings.Replace(c.BaseURL, "https://", "wss://", 1)
	u = strings.Replace(u, "http://", "ws://", 1)
	return u + "/api/websocket"
}

// connect opens an authenticated WebSocket connection.
func (c *Client) connect(ctx context.Context) (*websocket.Conn, error) {
	conn, _, err := websocket.Dial(ctx, c.wsURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("websocket dial: %w", err)
	}

	// Expect auth_required
	var msg map[string]any
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		conn.Close(websocket.StatusAbnormalClosure, "")
		return nil, fmt.Errorf("read auth_required: %w", err)
	}
	if msg["type"] != "auth_required" {
		conn.Close(websocket.StatusAbnormalClosure, "")
		return nil, fmt.Errorf("expected auth_required, got %v", msg["type"])
	}

	// Send auth
	auth := AuthMessage{Type: "auth", AccessToken: c.Token}
	if err := wsjson.Write(ctx, conn, auth); err != nil {
		conn.Close(websocket.StatusAbnormalClosure, "")
		return nil, fmt.Errorf("send auth: %w", err)
	}

	// Expect auth_ok or auth_invalid
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		conn.Close(websocket.StatusAbnormalClosure, "")
		return nil, fmt.Errorf("read auth response: %w", err)
	}
	if msg["type"] == "auth_invalid" {
		conn.Close(websocket.StatusAbnormalClosure, "")
		return nil, fmt.Errorf("authentication failed: invalid token")
	}
	if msg["type"] != "auth_ok" {
		conn.Close(websocket.StatusAbnormalClosure, "")
		return nil, fmt.Errorf("unexpected auth response: %v", msg["type"])
	}

	return conn, nil
}

// CallService calls a HA service via the WebSocket API.
func (c *Client) CallService(ctx context.Context, domain, service string, target *Target, data map[string]any) (*ResultMessage, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	msg := CallServiceMessage{
		ID:          c.nextID(),
		Type:        "call_service",
		Domain:      domain,
		Service:     service,
		ServiceData: data,
		Target:      target,
	}
	if err := wsjson.Write(ctx, conn, msg); err != nil {
		return nil, fmt.Errorf("send call_service: %w", err)
	}

	var result ResultMessage
	if err := wsjson.Read(ctx, conn, &result); err != nil {
		return nil, fmt.Errorf("read result: %w", err)
	}
	return &result, nil
}

// WatchStateChanges subscribes to state_changed events and calls fn for each.
// Blocks until ctx is cancelled or an error occurs.
func (c *Client) WatchStateChanges(ctx context.Context, entityFilter string, fn func(StateChangedData)) error {
	conn, err := c.connect(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	subID := c.nextID()
	sub := SubscribeMessage{
		ID:        subID,
		Type:      "subscribe_events",
		EventType: "state_changed",
	}
	if err := wsjson.Write(ctx, conn, sub); err != nil {
		return fmt.Errorf("send subscribe: %w", err)
	}

	// Read subscription confirmation
	var result ResultMessage
	if err := wsjson.Read(ctx, conn, &result); err != nil {
		return fmt.Errorf("read subscribe result: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("subscribe failed: %s", result.Error.Message)
	}

	for {
		var raw map[string]json.RawMessage
		if err := wsjson.Read(ctx, conn, &raw); err != nil {
			if ctx.Err() != nil {
				return nil // cancelled
			}
			return fmt.Errorf("read event: %w", err)
		}

		var evMsg EventMessage
		fullRaw, _ := json.Marshal(raw)
		if err := json.Unmarshal(fullRaw, &evMsg); err != nil {
			continue
		}
		if evMsg.Type != "event" {
			continue
		}

		var data StateChangedData
		if err := json.Unmarshal(evMsg.Event.Data, &data); err != nil {
			continue
		}

		if entityFilter != "" && data.EntityID != entityFilter {
			continue
		}

		fn(data)
	}
}
