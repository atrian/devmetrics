package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/atrian/devmetrics/internal/server/handlers"
	"github.com/atrian/devmetrics/internal/server/middlewares"
)

type Router struct {
	*chi.Mux
}

// registerMiddlewares общие middlewares для всех маршрутов
func registerMiddlewares(router *Router) {
	router.Mux.Use(middleware.RequestID)
	router.Mux.Use(middleware.Logger)
	router.Mux.Use(middlewares.GzipHandle)
}

func registerRoutes(router *Router, handler *handlers.Handler) {
	// По запросу GET http://<АДРЕС_СЕРВЕРА>/ сервер должен отдавать HTML-страничку со списком имён
	// и значений всех известных ему на текущий момент метрик.
	router.Get("/", handler.GetMetrics())

	// Сервер должен возвращать текущее значение запрашиваемой метрики в текстовом виде по запросу
	// GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> (со статусом http.StatusOK).
	// При попытке запроса неизвестной серверу метрики сервер должен возвращать http.StatusNotFound.
	router.Get("/value/{metricType}/{metricTitle}", handler.GetMetric())

	// Пинг соединения с БД
	router.Get("/ping", handler.GetPing())

	// Сохранение произвольных метрик,
	// POST /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	router.Post("/update/{metricType}/{metricTitle}/{metricValue}", handler.UpdateMetric())

	router.Post("/value/", handler.GetJSONMetric())
	// Пробуем анмаршалинг
	router.Post("/update/", handler.UpdateJSONMetric())
}

func New(handler *handlers.Handler) *Router {
	router := Router{
		Mux: chi.NewMux(),
	}

	// middlewares
	registerMiddlewares(&router)

	// routes
	registerRoutes(&router, handler)

	return &router
}
