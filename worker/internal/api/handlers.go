package api

import (
	"encoding/json"
	"net/http"

	"github.com/Alwanly/Houses-Prices/worker/internal/model"
	"go.uber.org/zap"
)

type listResponse struct {
	Items []*model.Listing `json:"items"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	listings, err := s.svc.GetListings(r.Context(), nil)
	if err != nil {
		s.logger.Error("get listings failed", zap.Error(err))
		http.Error(w, "failed to fetch listings", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(listResponse{Items: listings})
}

func (s *Server) handleTrigger(w http.ResponseWriter, r *http.Request) {
	site := r.URL.Query().Get("site")
	if site == "" {
		http.Error(w, "missing site param", http.StatusBadRequest)
		return
	}

	url := r.URL.Query().Get("url")

	go func() {
		if err := s.svc.ScrapeWebsite(r.Context(), site, url); err != nil {
			s.logger.Warn("background scrape failed", zap.Error(err))
			_ = s.notifier.NotifyError(r.Context(), site, err)
		} else {
			s.logger.Info("background scrape triggered", zap.String("site", site))
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}
