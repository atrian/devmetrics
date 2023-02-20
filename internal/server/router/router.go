// Package router сожержит все доступные веб роуты приложения
package router

import (
	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/atrian/devmetrics/internal/server/handlers"
	"github.com/atrian/devmetrics/internal/server/middlewares"
)

type Router struct {
	*chi.Mux
	config *serverconfig.Config
}

// RegisterMiddlewares общие middlewares для всех маршрутов
// Вызывать ДО регистрации маршрутов
func (r *Router) RegisterMiddlewares() *Router {
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middlewares.GzipHandle)

	return r
}

// RegisterCustomMiddlewares регистрация дополнительных middlewares.
// Вызывать ДО регистрации маршрутов
func (r *Router) RegisterCustomMiddlewares(middlewares []func(next http.Handler) http.Handler) *Router {
	r.Use(middlewares...)
	return r
}

// RegisterRoutes регистрация всех маршрутов бизнес логики приложения
// Вызывать ПОСЛЕ регистрации всех middlewares
func (r *Router) RegisterRoutes(handler *handlers.Handler) *Router {
	// Конфигурируем MW ограничения соединений для доверенных сетей
	trustedMW := middlewares.TrustedSubnetMW(r.config.Server.TrustedSubnet)

	r.Group(func(r chi.Router) {
		r.Use(trustedMW)
		// Сохранение произвольных метрик,
		// POST /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		r.Post("/update/{metricType}/{metricTitle}/{metricValue}", handler.UpdateMetric())

		// Получение 1 метрики в JSON
		r.Post("/value/", handler.GetJSONMetric())

		// Обновление 1 метрики в JSON
		r.Post("/update/", handler.UpdateJSONMetric())

		// Обновление пакета метрик из JSON
		r.Post("/updates/", handler.UpdateJSONMetrics())
	})

	// По запросу GET http://<АДРЕС_СЕРВЕРА>/ сервер должен отдавать HTML-страничку со списком имён
	// и значений всех известных ему на текущий момент метрик.
	r.Get("/", handler.GetMetrics())

	// Сервер должен возвращать текущее значение запрашиваемой метрики в текстовом виде по запросу
	// GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> (со статусом http.StatusOK).
	// При попытке запроса неизвестной серверу метрики сервер должен возвращать http.StatusNotFound.
	r.Get("/value/{metricType}/{metricTitle}", handler.GetMetric())

	// Пинг соединения с БД
	r.Get("/ping", handler.GetPing())

	return r
}

// New возвращает роутер со стандартной конфигурацией.
// Принимает слайс дополнительных кастомных middleware
func New(handler *handlers.Handler, middlewares []func(next http.Handler) http.Handler, config *serverconfig.Config) *Router {
	router := Router{
		Mux:    chi.NewMux(),
		config: config,
	}

	// middlewares
	router.RegisterMiddlewares()
	router.RegisterCustomMiddlewares(middlewares)

	// routes
	router.RegisterRoutes(handler)

	return &router
}
