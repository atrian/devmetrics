package handlers

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
)

func (h *Handler) GetPing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// соединение с БД
		dbPool, poolErr := pgxpool.Connect(context.Background(), h.config.Server.DBDSN)

		if poolErr != nil {
			http.Error(w, "Unable to connect to database:"+poolErr.Error(), http.StatusInternalServerError)
			return
		}

		if dbPool != nil {
			defer dbPool.Close()
			pingErr := dbPool.Ping(ctx)
			if pingErr != nil {
				http.Error(w, "Unable to ping database:"+pingErr.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
