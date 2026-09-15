// Package partnerdevice is the HTTP adapter for Partner Device admin CRUD
// and the live snapshot feed (REST + SSE) - see CONTEXT.md#environment and
// ADR 0011.
package partnerdevice

import (
	"encoding/json"
	"sync"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
)

// SSEHub fans out every freshly polled PartnerDeviceSnapshot to every
// connected SSE client. Implements ports/environment.PartnerDeviceBroadcaster.
// Channel-based (unlike the WebSocket Hub's direct conn.Write) because an
// SSE handler owns its connection via SendStreamWriter and needs something
// to range over.
type SSEHub struct {
	mu      sync.Mutex
	clients map[chan []byte]struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{clients: make(map[chan []byte]struct{})}
}

// Subscribe registers a new client and returns its feed. The channel is
// buffered so one slow client can't block Broadcast for everyone else -
// see Broadcast's non-blocking send.
func (h *SSEHub) Subscribe() chan []byte {
	ch := make(chan []byte, 8)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *SSEHub) Unsubscribe(ch chan []byte) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}

// Broadcast pushes the snapshot to every connected client. A client whose
// buffer is full has the update dropped rather than blocking every other
// subscriber - it will simply see the next tick's snapshot instead.
func (h *SSEHub) Broadcast(s environment.PartnerDeviceSnapshot) {
	payload, err := json.Marshal(toSnapshotResponse(s))
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- payload:
		default:
		}
	}
}
