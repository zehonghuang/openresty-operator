package httpapi

import (
	"encoding/json"
	"net/http"
	"openresty-operator/internal/controller"
)

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Handler defines a generic HTTP handler interface
type Handler interface {
	Serve(ctx context.Context, w http.ResponseWriter, r *http.Request)
	Path() string
}

// Server wraps the webhook server's mux and provides route registration
type Server struct {
	register func(path string, handler http.Handler) error
}

// NewServer initializes a new Server using the manager's metrics server
func NewServer(mgr manager.Manager) *Server {
	return &Server{
		register: mgr.AddMetricsServerExtraHandler,
	}
}

// RegisterHandler binds a generic handler into the metrics HTTP server
func (s *Server) RegisterHandler(h Handler) error {
	return s.register(h.Path(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, fmt.Sprintf("internal error: %v", err), http.StatusInternalServerError)
			}
		}()
		h.Serve(ctx, w, r)
	}))
}

type MetricsDNSCacheHandler struct{}

func (h *MetricsDNSCacheHandler) Path() string {
	return "/metrics/dns"
}

func (h *MetricsDNSCacheHandler) Serve(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(controller.DnsCache.Data)
}
