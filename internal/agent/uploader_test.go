package agent

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/atrian/devmetrics/internal/appconfig/agentconfig"
	"github.com/atrian/devmetrics/pkg/logger"
)

func TestUploader_buildStatUploadURL(t *testing.T) {
	type fields struct {
		client *http.Client
		config *agentconfig.Config
	}
	type args struct {
		metricType  string
		metricTitle string
		metricValue string
	}

	agentLogger := logger.NewZapLogger()
	config := agentconfig.NewConfig(agentLogger)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Basic func usage",
			fields: fields{
				client: nil,
				config: config,
			},
			args: args{
				metricType:  "gauge",
				metricTitle: "Alloc",
				metricValue: "0.0000",
			},
			want: fmt.Sprintf("%v://%v/updates/",
				config.Transport.Protocol,
				config.Transport.Address,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploader := &Uploader{
				HTTPClient: tt.fields.client,
				config:     tt.fields.config,
			}
			if got := uploader.buildStatsUploadURL(); got != tt.want {
				t.Errorf("buildStatsUploadURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
