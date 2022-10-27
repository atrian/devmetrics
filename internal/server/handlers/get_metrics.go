package handlers

import (
	"net/http"
	"text/template"
)

// GetMetrics получение всех сохраненных метрик в html формате GET /
func (h *Handler) GetMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html, err := template.ParseFiles(h.config.Server.MetricTemplateFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(http.StatusOK)
		html.Execute(w, h.storage.GetMetrics())
	}
}
