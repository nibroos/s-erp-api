// Package chat provides an in-memory websocket hub that tracks online users
// and fans chat events out to their live connections. It is intentionally
// decoupled from the persistence layer: it only moves already-serialized bytes.
package chat

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
)

// GlobalHub is the process-wide hub shared by the websocket handler and the
// REST controllers (so a message sent via REST still reaches online sockets).
var GlobalHub = NewHub()

// FanoutChannel is the Redis pub/sub channel used to broadcast chat frames
// across processes/instances (see EnableRedis). It lets a separate worker (the
// AI-reply consumer) deliver a message to whichever API instance holds the
// user's socket, and makes the hub multi-instance safe.
const FanoutChannel = "chat:fanout"

// fanoutMsg is the wire format on the Redis channel: the target user ids plus
// the already-serialized frame (Go encodes []byte as base64 in JSON).
type fanoutMsg struct {
	UserIDs []uint `json:"user_ids"`
	Payload []byte `json:"payload"`
}

// redisBroadcaster publishes fan-out frames to Redis instead of delivering them
// only to this process's local sockets.
type redisBroadcaster struct {
	rdb     *redis.Client
	channel string
}

// Client is a single websocket connection belonging to a user. Send is a
// buffered channel drained by the connection's write pump.
type Client struct {
	UserID uint
	Send   chan []byte
}

// Hub keeps the set of active clients per user id.
type Hub struct {
	mu          sync.RWMutex
	clients     map[uint]map[*Client]struct{}
	broadcaster *redisBroadcaster // nil = deliver only to local sockets
}

func NewHub() *Hub {
	return &Hub{clients: make(map[uint]map[*Client]struct{})}
}

// EnableRedis switches fan-out to Redis pub/sub. With subscribe=true (API
// instances) it also starts a subscriber that delivers frames received on the
// channel to this instance's local sockets; with subscribe=false (the AI-reply
// worker, which holds no sockets) it is publish-only. Enabling this makes the
// hub multi-instance safe and lets an out-of-process worker deliver replies.
func (h *Hub) EnableRedis(rdb *redis.Client, channel string, subscribe bool) {
	if rdb == nil {
		return
	}
	h.mu.Lock()
	h.broadcaster = &redisBroadcaster{rdb: rdb, channel: channel}
	h.mu.Unlock()
	if subscribe {
		go h.subscribe(rdb, channel)
	}
}

// subscribe reads the fan-out channel and delivers each frame to local sockets.
func (h *Hub) subscribe(rdb *redis.Client, channel string) {
	sub := rdb.Subscribe(context.Background(), channel)
	log.Printf("chat hub: subscribed to Redis fan-out channel %q", channel)
	for msg := range sub.Channel() {
		var fm fanoutMsg
		if err := json.Unmarshal([]byte(msg.Payload), &fm); err != nil {
			continue
		}
		h.deliverLocal(fm.UserIDs, fm.Payload)
	}
}

// Add registers a client connection for its user. It returns true when this is
// the user's first live connection (an offline -> online transition).
func (h *Hub) Add(c *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	becameOnline := len(h.clients[c.UserID]) == 0
	if h.clients[c.UserID] == nil {
		h.clients[c.UserID] = make(map[*Client]struct{})
	}
	h.clients[c.UserID][c] = struct{}{}
	return becameOnline
}

// Remove deregisters a client connection and closes its send channel. It
// returns true when the user has no remaining connections (an online ->
// offline transition).
func (h *Hub) Remove(c *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	set, ok := h.clients[c.UserID]
	if !ok {
		return false
	}
	if _, exists := set[c]; exists {
		delete(set, c)
		close(c.Send)
	}
	if len(set) == 0 {
		delete(h.clients, c.UserID)
		return true
	}
	return false
}

// FilterOnline returns the subset of the given user ids that are currently online.
func (h *Hub) FilterOnline(userIDs []uint) []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()
	online := make([]uint, 0, len(userIDs))
	for _, id := range userIDs {
		if len(h.clients[id]) > 0 {
			online = append(online, id)
		}
	}
	return online
}

// IsOnline reports whether a user has at least one live connection.
func (h *Hub) IsOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID]) > 0
}

// SendToUsers delivers payload to the given users' live connections. When Redis
// fan-out is enabled it publishes the frame to the shared channel (every
// instance's subscriber then delivers it to its own sockets); otherwise, or if
// the publish fails, it delivers to this process's local sockets directly.
func (h *Hub) SendToUsers(userIDs []uint, payload []byte) {
	h.mu.RLock()
	b := h.broadcaster
	h.mu.RUnlock()

	if b != nil {
		if data, err := json.Marshal(fanoutMsg{UserIDs: userIDs, Payload: payload}); err == nil {
			if err := b.rdb.Publish(context.Background(), b.channel, data).Err(); err == nil {
				return // the subscriber(s) will deliver to local sockets
			}
			// Publish failed — fall back to local delivery so at least sockets
			// on this instance still get the frame.
		}
	}
	h.deliverLocal(userIDs, payload)
}

// deliverLocal fans a frame out to this process's live connections only. It is
// best-effort and non-blocking: a client whose buffer is full is skipped rather
// than stalling the whole fan-out.
func (h *Hub) deliverLocal(userIDs []uint, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, uid := range userIDs {
		for c := range h.clients[uid] {
			select {
			case c.Send <- payload:
			default:
				// Slow consumer; drop this frame for that connection.
			}
		}
	}
}
