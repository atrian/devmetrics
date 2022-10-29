package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func (h *Handler) GetPing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// соединение с БД
		dbPool, poolErr := pgxpool.Connect(context.Background(), h.config.Server.DBDSN)

		if poolErr != nil {
			http.Error(w, "Unable to connect to database:"+poolErr.Error(), http.StatusInternalServerError)
			return
		}

		if dbPool != nil {
			defer dbPool.Close()

			ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
			defer cancel()
			pingErr := dbPool.Ping(ctx) // наследуем контекcт запроса r *http.Request, добавляем таймаут
			if pingErr != nil {
				http.Error(w, "Unable to ping database:"+pingErr.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}