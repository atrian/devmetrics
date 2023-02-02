// Package handlers Хендлеры сервиса, см. документацию в Swagger
package handlers

import (
	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/crypter"
	"github.com/atrian/devmetrics/internal/server/storage"
	"github.com/atrian/devmetrics/internal/signature"
	"github.com/atrian/devmetrics/pkg/logger"
)

type Handler struct {
	storage storage.Repository
	config  *serverconfig.Config
	hasher  signature.Hasher // hasher для проверки подписи метрик
	crypter crypter.Crypter  // crypter для расшифровки метрик приватным ключом
	logger  logger.Logger
}

func New(config *serverconfig.Config, storage storage.Repository, logger logger.Logger) *Handler {
	h := &Handler{
		storage: storage,
		config:  config,
		hasher:  signature.NewSha256Hasher(),
		logger:  logger,
	}

	// Конфигурируем модуль расшифровки метрик, установлен ключ если он есть
	km := crypter.New()
	if h.config.Server.CryptoKey != "" {
		secret, err := km.ReadPrivateKey(h.config.Server.CryptoKey)
		if err != nil {
			h.logger.Error("Can't read private key", err)
		}
		km.RememberPrivateKey(secret)
	}
	h.crypter = km

	return h
}
