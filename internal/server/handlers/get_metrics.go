package handlers

import (
	"net/http"
	"text/template"
)

// GetMetrics получение всех сохраненных метрик в html формате GET /
//
//	@Tags Metrics
//	@Summary Выводит все метрики в html виде
//	@Produce html
//	@Success 200 {string} string "HTML страница с метриками"
//	@Failure 500
//	@Router / [get]
func (h *Handler) GetMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html, err := template.ParseFiles(h.config.Server.MetricTemplateFile)
		if err != nil {
			h.logger.Error("GetMetrics template.ParseFiles", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/html")
		w.WriteHeader(http.StatusOK)
		htmlErr := html.Execute(w, h.storage.GetMetrics())
		if htmlErr != nil {
			h.logger.Error("html.Execute error", htmlErr)
		}
	}
}
