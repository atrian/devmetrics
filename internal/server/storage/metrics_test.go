package storage

import (
	"testing"

	"github.com/atrian/devmetrics/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
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
	suite.storage.StoreCounter("PollCount", int64(1))

	value, exist := suite.storage.GetCounter("PollCount")
	assert.Equal(suite.T(), true, exist)
	assert.Equal(suite.T(), int64(1), value)

	suite.storage.StoreCounter("PollCount", int64(2022))
	// при последующем сохранении значение увеличилось а не перезаписалось 2022 + 1
	value, _ = suite.storage.GetCounter("PollCount")
	assert.Equal(suite.T(), int64(2023), value)
}

func (suite *HandlersTestSuite) TestStorage_StoreGauge() {
	suite.storage.StoreGauge("Alloc", float64(1))

	// значение сохранилось в мапе по ключу Alloc
	val, exist := suite.storage.GetGauge("Alloc")
	assert.Equal(suite.T(), true, exist)
	assert.Equal(suite.T(), float64(1), val)

	suite.storage.StoreGauge("Alloc", float64(777))

	// при последующем сохранении значение перезаписалось
	val, _ = suite.storage.GetGauge("Alloc")
	assert.Equal(suite.T(), float64(777), val)
}

// Для запуска через Go test
func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}
