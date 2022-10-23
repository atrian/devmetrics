package agent

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/atrian/devmetrics/internal/appconfig/agentConfig"
)

func TestUploader_buildStatUploadURL(t *testing.T) {
	type fields struct {
		client *http.Client
		config *agentConfig.HTTPConfig
	}
	type args struct {
		metricType  string
		metricTitle string
		metricValue string
	}

	config := agentConfig.NewConfig()

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
				config: &config.HTTP,
			},
			args: args{
				metricType:  "gauge",
				metricTitle: "Alloc",
				metricValue: "0.0000",
			},
			want: fmt.Sprintf("%v://%v/update/",
				config.HTTP.Protocol,
				config.HTTP.Address,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploader := &Uploader{
				client: tt.fields.client,
				config: tt.fields.config,
			}
			if got := uploader.buildStatUploadURL(); got != tt.want {
				t.Errorf("buildStatUploadURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
