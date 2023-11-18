package log

import (
	"context"
	"testing"

	"github.com/kcloutie/knot/pkg/config"
	"github.com/kcloutie/knot/pkg/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestProvider_SendNotification(t *testing.T) {
	type args struct {
		data         *message.NotificationData
		notification config.Notification
	}
	tests := []struct {
		name    string
		v       *Provider
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			v:    New(),
			args: args{
				data: &message.NotificationData{
					Data: map[string]interface{}{
						"prop1": "world",
					},
					Attributes: map[string]string{},
					ID:         "1",
				},
				notification: config.Notification{
					Properties: map[string]string{
						"message": "hello {{ .data.prop1 }}",
					},
				},
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			observedZapCore, observedLogs := observer.New(zap.InfoLevel)
			observedLogger := zap.New(observedZapCore)
			tt.v.SetLogger(observedLogger)
			tt.v.SetNotification(tt.args.notification)

			if err := tt.v.SendNotification(context.Background(), tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Provider.SendNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, 1, observedLogs.Len())
			firstLog := observedLogs.All()[0]
			assert.Equal(t, "hello world", firstLog.Message)
		})
	}
}
