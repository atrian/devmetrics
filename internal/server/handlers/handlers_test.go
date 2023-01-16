package handlers_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/server/handlers"
	"github.com/atrian/devmetrics/internal/server/router"
	"github.com/atrian/devmetrics/internal/server/storage"
	"github.com/atrian/devmetrics/pkg/logger"
)

type HandlersTestSuite struct {
	suite.Suite
	config  *serverconfig.Config
	storage storage.IRepository
	router  *router.Router
	logger  logger.Logger
}

func (suite *HandlersTestSuite) SetupSuite() {
	suite.logger = logger.NewZapLogger()
	suite.config = serverconfig.NewServerConfig(suite.logger)
	suite.storage = storage.NewMemoryStorage(suite.config, suite.logger)
	suite.router = router.New(handlers.New(suite.config, suite.storage, suite.logger))
}

func (suite *HandlersTestSuite) TestUpdateHandlers() {
	ts := httptest.NewServer(suite.router)
	defer ts.Close()

	// тестовые кейсы из прошлых инкрементов
	tt := []struct {
		testName     string
		method       string
		endpoint     string
		expectedBody string
		statusCode   int
	}{
		{
			"Counter metric with valid params",
			"POST",
			"/update/counter/PollCount/3",
			"3",
			http.StatusOK,
		}, {
			"Counter metric with invalid params",
			"POST",
			"/update/counter/PollCount/3+",
			"Cant store metric\n",
			http.StatusBadRequest,
		}, {
			"Gauge metric with valid params",
			"POST",
			"/update/gauge/RandomValue/3.0402",
			"3.0402",
			http.StatusOK,
		}, {
			"Not implemented metric request",
			"POST",
			"/update/gaugeInvalid/InvalidMetrics/6",
			"Not implemented\n",
			http.StatusNotImplemented,
		}, {
			"Request without ID",
			"POST",
			"/update/gauge/",
			"404 page not found\n",
			http.StatusNotFound,
		},
	}

	for _, tc := range tt {
		suite.T().Run(tc.testName, func(t *testing.T) {
			statusCode, body := testRequest(t, ts, tc.method, tc.endpoint)
			assert.Equal(t, tc.statusCode, statusCode)
			assert.Equal(t, tc.expectedBody, body)
		})
	}
}

func (suite *HandlersTestSuite) TestUpdateCounterInSeries() {
	ts := httptest.NewServer(suite.router)
	defer ts.Close()

	// последовательное сохранение значения счетчика
	_, _ = testRequest(suite.T(), ts, "POST", "/update/counter/TestCounter/5")
	_, _ = testRequest(suite.T(), ts, "POST", "/update/counter/TestCounter/8")
	statusCode, body := testRequest(suite.T(), ts, "GET", "/value/counter/TestCounter")
	assert.Equal(suite.T(), http.StatusOK, statusCode)
	assert.Equal(suite.T(), "13", body)
}

// Для запуска через Go test
func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

// testRequest вспомогательный метод для отправки запросов
func testRequest(t *testing.T, ts *httptest.Server, method, path string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()
	return resp.StatusCode, string(respBody)
}

func ExampleHandler_UpdateMetric() {
	// Подготавливаем все зависимости, логгер, конфигурацию приложения, хранилище (In Memory) и роутер
	appLogger := logger.NewZapLogger()
	appConf := serverconfig.NewServerConfigWithoutFlags(appLogger)
	memStorage := storage.NewMemoryStorage(appConf, appLogger)
	r := router.New(handlers.New(appConf, memStorage, appLogger))

	// Запускаем тестовый сервер
	testServer := httptest.NewServer(r)
	defer testServer.Close()

	// Запрос отправляется на Endpoint /update/counter/PollCount/3 POST методом
	endpoint := testServer.URL + "/update/counter/PollCount/3"

	request, _ := http.NewRequest(http.MethodPost, endpoint, nil)
	response, respErr := http.DefaultClient.Do(request)
	if respErr != nil {
		appLogger.Fatal("http.DefaultClient.Do error", respErr)
	}
	defer response.Body.Close()

	responseBody, _ := io.ReadAll(response.Body)

	// В случае успеха сервис отвечает кодом 200 и текущим значением метрики
	fmt.Println(response.StatusCode, string(responseBody))

	// Output:
	// 200 3
}

func ExampleHandler_UpdateJSONMetrics() {
	// Подготавливаем к передаче метрики и счетчики в JSON формате
	metricsInJSON := `[{"id":"YourCounter","type":"counter","delta":555},{"id":"YourGauge","type":"gauge","value":100.500}]`
	metricsReader := strings.NewReader(metricsInJSON)

	// Подготавливаем все зависимости, логгер, конфигурацию приложения, хранилище (In Memory) и роутер
	appLogger := logger.NewZapLogger()
	appConf := serverconfig.NewServerConfigWithoutFlags(appLogger)
	memStorage := storage.NewMemoryStorage(appConf, appLogger)
	r := router.New(handlers.New(appConf, memStorage, appLogger))

	// Запускаем тестовый сервер
	testServer := httptest.NewServer(r)
	defer testServer.Close()

	// Запрос отправляется на Endpoint /updates/ POST методом. В теле передается JSON с метриками
	endpoint := testServer.URL + "/updates/"

	request, _ := http.NewRequest(http.MethodPost, endpoint, metricsReader)
	response, respErr := http.DefaultClient.Do(request)
	if respErr != nil {
		appLogger.Fatal("http.DefaultClient.Do error", respErr)
	}
	defer response.Body.Close()

	responseBody, _ := io.ReadAll(response.Body)

	// В случае успеха сервис отвечает кодом 200 и JSON содержащим статус и текущие значения переданных метрик
	fmt.Println(response.StatusCode, string(responseBody))

	// Output:
	// 200 {"Status":"OK","Updated":[{"id":"YourCounter","type":"counter","delta":555},{"id":"YourGauge","type":"gauge","value":100.5}]}
}

func ExampleHandler_GetJSONMetric() {
	// Предварительная загрузка данных в сервис
	metricsInJSON := `[{"id":"YourCounter","type":"counter","delta":876},{"id":"YourGauge","type":"gauge","value":100.500}]`
	metricsReader := strings.NewReader(metricsInJSON)

	// Подготавливаем все зависимости, логгер, конфигурацию приложения, хранилище (In Memory) и роутер
	appLogger := logger.NewZapLogger()
	appConf := serverconfig.NewServerConfigWithoutFlags(appLogger)
	memStorage := storage.NewMemoryStorage(appConf, appLogger)
	r := router.New(handlers.New(appConf, memStorage, appLogger))

	// Запускаем тестовый сервер
	testServer := httptest.NewServer(r)
	defer testServer.Close()

	// Отправляем запрос на сохранение предварительно подготовленных данных
	// на Endpoint /updates/ POST методом. В теле передается JSON с метриками
	endpoint := testServer.URL + "/updates/"

	request, _ := http.NewRequest(http.MethodPost, endpoint, metricsReader)
	response, respErr := http.DefaultClient.Do(request)
	if respErr != nil {
		appLogger.Fatal("http.DefaultClient.Do error", respErr)
	}
	response.Body.Close()

	// Делаем запрос нужной метрики
	// на Endpoint /value/ POST методом
	// id - имя метрики или счетчика
	// type - тип метрики: gauge, counter

	weWant := `{"id":"YourCounter","type":"counter"}`
	metricReader := strings.NewReader(weWant)

	endpoint = testServer.URL + "/value/"
	request, _ = http.NewRequest(http.MethodPost, endpoint, metricReader)
	response, respErr = http.DefaultClient.Do(request)
	if respErr != nil {
		appLogger.Fatal("http.DefaultClient.Do error", respErr)
	}
	defer response.Body.Close()

	responseBody, _ := io.ReadAll(response.Body)

	// В случае успеха сервис отвечает кодом 200 и текущим значением метрики в формате JSON
	fmt.Println(response.StatusCode, string(responseBody))

	// Output:
	// 200 {"id":"YourCounter","type":"counter","delta":876}
}

func ExampleHandler_GetMetric() {
	// Предварительная загрузка данных в сервис
	metricsInJSON := `[{"id":"YourCounter","type":"counter","delta":741},{"id":"YourGauge","type":"gauge","value":500.500}]`
	metricsReader := strings.NewReader(metricsInJSON)

	// Подготавливаем все зависимости, логгер, конфигурацию приложения, хранилище (In Memory) и роутер
	appLogger := logger.NewZapLogger()
	appConf := serverconfig.NewServerConfigWithoutFlags(appLogger)
	memStorage := storage.NewMemoryStorage(appConf, appLogger)
	r := router.New(handlers.New(appConf, memStorage, appLogger))

	// Запускаем тестовый сервер
	testServer := httptest.NewServer(r)
	defer testServer.Close()

	// Отправляем запрос на сохранение предварительно подготовленных данных
	// на Endpoint /updates/ POST методом. В теле передается JSON с метриками
	endpoint := testServer.URL + "/updates/"

	request, _ := http.NewRequest(http.MethodPost, endpoint, metricsReader)
	response, respErr := http.DefaultClient.Do(request)
	if respErr != nil {
		appLogger.Fatal("http.DefaultClient.Do error", respErr)
	}
	response.Body.Close()

	// Делаем запрос нужной метрики
	// на Endpoint /value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> GET методом
	// тип метрики: gauge, counter

	weWant := `{"id":"YourCounter","type":"counter"}`
	metricReader := strings.NewReader(weWant)

	endpoint = testServer.URL + "/value/gauge/YourGauge"
	request, _ = http.NewRequest(http.MethodGet, endpoint, metricReader)
	response, respErr = http.DefaultClient.Do(request)
	if respErr != nil {
		appLogger.Fatal("http.DefaultClient.Do error", respErr)
	}
	defer response.Body.Close()

	responseBody, _ := io.ReadAll(response.Body)

	// В случае успеха сервис отвечает кодом 200 и текущим значением метрики
	fmt.Println(response.StatusCode, string(responseBody))

	// Output:
	// 200 500.5
}

func ExampleHandler_UpdateJSONMetric() {
	// Подготавливаем к передаче метрику или счетчик в JSON формате
	metricInJSON := `{"id":"YourCounter","type":"counter","delta":835}`
	metricReader := strings.NewReader(metricInJSON)

	// Подготавливаем все зависимости, логгер, конфигурацию приложения, хранилище (In Memory) и роутер
	appLogger := logger.NewZapLogger()
	appConf := serverconfig.NewServerConfigWithoutFlags(appLogger)
	memStorage := storage.NewMemoryStorage(appConf, appLogger)
	r := router.New(handlers.New(appConf, memStorage, appLogger))

	// Запускаем тестовый сервер
	testServer := httptest.NewServer(r)
	defer testServer.Close()

	// Запрос отправляется на Endpoint /update/ POST методом. В теле передается JSON с метрикой
	endpoint := testServer.URL + "/update/"

	request, _ := http.NewRequest(http.MethodPost, endpoint, metricReader)
	response, respErr := http.DefaultClient.Do(request)
	if respErr != nil {
		appLogger.Fatal("http.DefaultClient.Do error", respErr)
	}
	defer response.Body.Close()

	responseBody, _ := io.ReadAll(response.Body)

	// В случае успеха сервис отвечает кодом 200 и JSON содержащим статус и текущие значения переданной метрики
	fmt.Println(response.StatusCode, string(responseBody))

	// Output:
	// 200 {"id":"YourCounter","type":"counter","delta":835}
}
