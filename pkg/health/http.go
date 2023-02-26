package health

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/g4s8/go-lifecycle/pkg/lifecycle"
	"github.com/g4s8/go-lifecycle/pkg/types"
)

var _ http.Handler = (*handler)(nil)

type serviceState struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type healthState struct {
	Healthy  bool           `json:"healthy"`
	Services []serviceState `json:"services"`
}

type handler struct {
	states   []lifecycle.ServiceState
	statesMx sync.RWMutex
}

func (h *handler) update(states []lifecycle.ServiceState) {
	h.statesMx.Lock()
	defer h.statesMx.Unlock()
	h.states = states
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	states := make([]serviceState, len(h.states))
	healthy := true
	h.statesMx.RLock()
	for i, st := range h.states {
		states[i] = serviceState{
			ID:     st.ID,
			Name:   st.Name,
			Status: st.Status.String(),
		}
		if st.Error != nil {
			states[i].Error = st.Error.Error()
		}
		if st.Status == types.ServiceStatusError {
			healthy = false
		}
	}
	h.statesMx.RUnlock()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if !healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(healthState{
		Healthy:  healthy,
		Services: states,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
