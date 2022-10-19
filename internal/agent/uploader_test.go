package agent

import (
	"fmt"
	"github.com/atrian/devmetrics/internal/appconfig"
	"net/http"
	"testing"
)

func TestUploader_buildStatUploadURL(t *testing.T) {
	type fields struct {
		client *http.Client
		config *appconfig.HTTPConfig
	}
	type args struct {
		metricType  string
		metricTitle string
		metricValue string
	}

	config := appconfig.NewConfig()

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
			want: fmt.Sprintf("http://%v:%d/update/",
				config.HTTP.Server,
				config.HTTP.Port,
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
