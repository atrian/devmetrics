package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/pkg/logger"
)

type HandlersTestSuite struct {
	suite.Suite
	config  *serverconfig.Config
	storage Repository
	logger  logger.Logger
}

func (suite *HandlersTestSuite) SetupSuite() {
	suite.logger = logger.NewZapLogger()
	suite.config = serverconfig.NewServerConfig(suite.logger)
}

func (suite *HandlersTestSuite) SetupTest() {
	suite.storage = NewMemoryStorage(suite.config, suite.logger)
}

func (suite *HandlersTestSuite) TestStorage_StoreCounter() {
	err := suite.storage.StoreCounter("PollCount", int64(1))
	if err != nil {
		return
	}

	value, exist := suite.storage.GetCounter("PollCount")
	assert.Equal(suite.T(), true, exist)
	assert.Equal(suite.T(), int64(1), value)

	err = suite.storage.StoreCounter("PollCount", int64(2022))
	if err != nil {
		return
	}
	// при последующем сохранении значение увеличилось а не перезаписалось 2022 + 1
	value, _ = suite.storage.GetCounter("PollCount")
	assert.Equal(suite.T(), int64(2023), value)
}

func (suite *HandlersTestSuite) TestStorage_StoreGauge() {
	err := suite.storage.StoreGauge("Alloc", float64(1))
	if err != nil {
		return
	}

	// значение сохранилось в мапе по ключу Alloc
	val, exist := suite.storage.GetGauge("Alloc")
	assert.Equal(suite.T(), true, exist)
	assert.Equal(suite.T(), float64(1), val)

	err = suite.storage.StoreGauge("Alloc", float64(777))
	if err != nil {
		return
	}

	// при последующем сохранении значение перезаписалось
	val, _ = suite.storage.GetGauge("Alloc")
	assert.Equal(suite.T(), float64(777), val)
}

// Для запуска через Go test
func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}
