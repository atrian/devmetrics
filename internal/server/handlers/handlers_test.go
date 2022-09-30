package handlers

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// тест сделан по мотивам ролика justforfunc #16: unit testing HTTP servers
func TestUpdateMetricHandler(t *testing.T) {
	tt := []struct {
		testName     string
		endpoint     string
		expectedBody string
		statusCode   int
	}{
		{
			"Counter metric with valid params",
			"http://localhost:8080/update/counter/PollCount/3",
			"3",
			http.StatusOK,
		}, {
			"Counter metric with invalid params",
			"http://localhost:8080/update/counter/PollCount/3+",
			"Cant store metric\n",
			http.StatusBadRequest,
		}, {
			"Gauge metric with valid params",
			"http://localhost:8080/update/gauge/RandomValue/3.0402",
			"3.0402",
			http.StatusOK,
		}, {
			"InvalidMetrics metric request",
			"http://localhost:8080/update/gauge/InvalidMetrics/6",
			"Cant store metric\n",
			http.StatusBadRequest,
		}, {
			"InvalidMetrics metric types",
			"http://localhost:8080/update/invalidtype/InvalidMetrics/6",
			"Can't validate update request\n",
			http.StatusBadRequest,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, tc.endpoint, nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}

			// создаем экземпляр хендлера и рекордер ответа сервера
			handler := NewUpdateMetricHandler()
			responseRecorder := httptest.NewRecorder()
			handler.UpdateMetric(responseRecorder, request)
			response := responseRecorder.Result()

			assert.Equal(t, tc.statusCode, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("Could not read response body: %v", err)
			}
			defer response.Body.Close()

			assert.Equal(t, tc.expectedBody, string(body))
		})
	}
}

// Тестируем суммирование счетчика CounterMetric
func TestUpdateCounterMetricAccumulatingValues(t *testing.T) {
	// Первый запрос на обновление метрики счетчика
	request, err := http.NewRequest(http.MethodPost, "http://localhost:8080/update/counter/PollCount/3", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// создаем экземпляр хендлера и рекордер ответа сервера
	handler := NewUpdateMetricHandler()
	responseRecorder := httptest.NewRecorder()
	handler.UpdateMetric(responseRecorder, request)
	response := responseRecorder.Result()

	firtsResponseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Could not read response body: %v", err)
	}
	defer response.Body.Close()
	assert.Equal(t, "3", string(firtsResponseBody))

	// второй запрос на обновление метрики счетчика
	request2, err := http.NewRequest(http.MethodPost, "http://localhost:8080/update/counter/PollCount/4", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	secondRecorder := httptest.NewRecorder()
	handler.UpdateMetric(secondRecorder, request2)
	secondResponse := secondRecorder.Result()

	secondResponseBody, err := io.ReadAll(secondResponse.Body)
	if err != nil {
		t.Fatalf("Could not read response body: %v", err)
	}
	defer secondResponse.Body.Close()

	assert.Equal(t, "7", string(secondResponseBody))
}

// Рыба этого теста сгенерирована через Golang
func Test_endpointParser(t *testing.T) {
	type args struct {
		endpoint string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Base endpointParser usage",
			args: args{endpoint: "/update/counter/PollCount/3"},
			want: []string{"update", "counter", "PollCount", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := endpointParser(tt.args.endpoint); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("endpointParser() = %v, want %v", got, tt.want)
			}
		})
	}
}
