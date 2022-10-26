package handlers

import (
	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/server/storage"
)

type Handler struct {
	storage storage.Repository
	config  *serverconfig.Config
}

func New(config *serverconfig.Config, storage storage.Repository) *Handler {
	h := &Handler{
		storage: storage,
		config:  config,
	}

	return h
}
