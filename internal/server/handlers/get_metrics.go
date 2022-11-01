package handlers

import (
	"net/http"
	"text/template"

	"go.uber.org/zap"
)

// GetMetrics получение всех сохраненных метрик в html формате GET /
func (h *Handler) GetMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html, err := template.ParseFiles(h.config.Server.MetricTemplateFile)
		if err != nil {
			h.logger.Error("GetMetrics template.ParseFiles", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(http.StatusOK)
		html.Execute(w, h.storage.GetMetrics())
	}
}
