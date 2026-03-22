package tunnel

import (
	"sync"
	"time"
)

// Session은 활성 터널 세션을 나타낸다.
type Session struct {
	ID          string
	Subdomain   string
	AuthToken   string
	LocalPort   int
	ConnectedAt time.Time
	LastSeen    time.Time
}

// PendingRequest는 CLI 응답을 기다리는 HTTP 요청이다.
type PendingRequest struct {
	ch chan *TunnelResponse
}

func newPendingRequest() *PendingRequest {
	return &PendingRequest{ch: make(chan *TunnelResponse, 1)}
}

// Registry는 pendingRequests를 스레드-안전하게 관리한다.
type Registry struct {
	mu       sync.RWMutex
	pending  map[string]*PendingRequest
}

func NewRegistry() *Registry {
	return &Registry{pending: make(map[string]*PendingRequest)}
}

func (r *Registry) Add(id string) *PendingRequest {
	pr := newPendingRequest()
	r.mu.Lock()
	r.pending[id] = pr
	r.mu.Unlock()
	return pr
}

func (r *Registry) Deliver(id string, resp *TunnelResponse) bool {
	r.mu.RLock()
	pr, ok := r.pending[id]
	r.mu.RUnlock()
	if !ok {
		return false
	}
	pr.ch <- resp
	return true
}

func (r *Registry) Remove(id string) {
	r.mu.Lock()
	delete(r.pending, id)
	r.mu.Unlock()
}
