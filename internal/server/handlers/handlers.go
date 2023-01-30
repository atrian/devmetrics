// Package handlers Хендлеры сервиса, см. документацию в Swagger
package handlers

import (
	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/server/storage"
	"github.com/atrian/devmetrics/internal/signature"
	"github.com/atrian/devmetrics/pkg/logger"
)

type Handler struct {
	storage storage.Repository
	config  *serverconfig.Config
	hasher  signature.Hasher
	logger  logger.Logger
}

func New(config *serverconfig.Config, storage storage.Repository, logger logger.Logger) *Handler {
	h := &Handler{
		storage: storage,
		config:  config,
		hasher:  signature.NewSha256Hasher(),
		logger:  logger,
	}

	return h
}
