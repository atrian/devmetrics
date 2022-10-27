package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/atrian/devmetrics/internal/appconfig/serverconfig"
	"github.com/atrian/devmetrics/internal/server/handlers"
	"github.com/atrian/devmetrics/internal/server/router"
	"github.com/atrian/devmetrics/internal/server/storage"
)

type HandlersTestSuite struct {
	suite.Suite
	config  *serverconfig.Config
	storage storage.Repository
	router  *router.Router
}

func (suite *HandlersTestSuite) SetupSuite() {
	suite.config = serverconfig.NewServerConfig()
	suite.storage = storage.NewMemoryStorage(suite.config)
	suite.router = router.New(handlers.New(suite.config, suite.storage))
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

// Ctrl-c Ctrl-v из учебника практикума.
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
