package handlers

import (
	"go.uber.org/zap"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/crypto"
	"github.com/atrian/devmetrics/internal/server/storage"
)

type Handler struct {
	storage storage.Repository
	config  *serverconfig.Config
	hasher  crypto.Hasher
	logger  *zap.Logger
}

func New(config *serverconfig.Config, storage storage.Repository, logger *zap.Logger) *Handler {
	h := &Handler{
		storage: storage,
		config:  config,
		hasher:  crypto.NewSha256Hasher(),
		logger:  logger,
	}

	return h
}
