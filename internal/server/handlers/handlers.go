// Package handlers Хендлеры сервиса, см. документацию в Swagger
package handlers

import (
	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/crypto"
	"github.com/atrian/devmetrics/internal/server/storage"
	"github.com/atrian/devmetrics/pkg/logger"
)

type Handler struct {
	storage storage.IRepository
	config  *serverconfig.Config
	hasher  crypto.Hasher
	logger  logger.Logger
}

func New(config *serverconfig.Config, storage storage.IRepository, logger logger.Logger) *Handler {
	h := &Handler{
		storage: storage,
		config:  config,
		hasher:  crypto.NewSha256Hasher(),
		logger:  logger,
	}

	return h
}
