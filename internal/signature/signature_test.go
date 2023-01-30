package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasher(t *testing.T) {
	h := NewSha256Hasher()
	validKey := "my_test_key"
	invalidKey := "invalid_key"

	tt := []struct {
		testName            string
		agentKey            string
		serverKey           string
		expectedCheckStatus bool
	}{
		{
			testName:            "Agent and server key is same",
			agentKey:            validKey,
			serverKey:           validKey,
			expectedCheckStatus: true,
		},
		{
			testName:            "Agent and server key is NOT same",
			agentKey:            invalidKey,
			serverKey:           validKey,
			expectedCheckStatus: false,
		},
		{
			testName:            "Agent dont have a key",
			agentKey:            "",
			serverKey:           validKey,
			expectedCheckStatus: false,
		},
		{
			testName:            "Server dont have a key",
			agentKey:            validKey,
			serverKey:           "",
			expectedCheckStatus: false,
		},
		{
			testName:            "Server and agent dont have a key",
			agentKey:            "",
			serverKey:           "",
			expectedCheckStatus: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testName, func(t *testing.T) {
			metric := "test_metric_string"
			agentSign := h.Hash(metric, tc.agentKey)
			assert.True(t, tc.expectedCheckStatus == h.Compare(agentSign, metric, tc.serverKey))
		})
	}
}
