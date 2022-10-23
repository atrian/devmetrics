package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/server/storage"
)

type Handler struct {
	*chi.Mux
	storage storage.Repository
	config  *serverconfig.Config
}

func NewHandler(config *serverconfig.Config, storage storage.Repository) *Handler {
	h := &Handler{
		Mux:     chi.NewMux(),
		storage: storage,
		config:  config,
	}

	// регистрируем middlewares
	h.Mux.Use(middleware.RequestID)
	h.Mux.Use(middleware.Logger)

	// gzip
	h.Mux.Use(middleware.Compress(5))

	// По запросу GET http://<АДРЕС_СЕРВЕРА>/ сервер должен отдавать HTML-страничку со списком имён
	// и значений всех известных ему на текущий момент метрик.
	h.Get("/", h.GetMetrics())

	// Сервер должен возвращать текущее значение запрашиваемой метрики в текстовом виде по запросу
	// GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> (со статусом http.StatusOK).
	// При попытке запроса неизвестной серверу метрики сервер должен возвращать http.StatusNotFound.
	h.Get("/value/{metricType}/{metricTitle}", h.GetMetric())

	// Сохранение произвольных метрик,
	// POST /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	h.Post("/update/{metricType}/{metricTitle}/{metricValue}", h.UpdateMetric())

	h.Post("/value/", h.GetJSONMetric())
	// Пробуем анмаршалинг
	h.Post("/update/", h.UpdateJSONMetric())

	return h
}
