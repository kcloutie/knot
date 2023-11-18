package pubsub

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/kcloutie/knot/pkg/message"
	"go.uber.org/zap/zaptest"
)

func TestListener_ParsePayload(t *testing.T) {
	goodPayload := pubsub.Message{
		Attributes: map[string]string{
			"att1": "att1Val",
		},
		Data: []byte(`{"test":"123"}`),
		ID:   "1",
	}
	goodPayloadBytes, _ := json.Marshal(goodPayload)
	expectedGood, _ := message.ToNotificationData(&goodPayload)
	type args struct {
		payload []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *message.NotificationData
		wantErr string
	}{
		{
			name: "non json payload",
			args: args{
				payload: []byte(`dude`),
			},
			want:    nil,
			wantErr: "Failed to unmarshal body to the pubsub.Message type. Error: invalid character 'd' looking for beginning of value",
		},
		{
			name: "bad payload",
			args: args{
				payload: []byte(`{}`),
			},
			want:    nil,
			wantErr: "failed to unmarshal the pub/sub data property of the message - unexpected end of JSON input",
		},
		{
			name: "success",
			args: args{
				payload: goodPayloadBytes,
			},
			want:    &expectedGood,
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger := zaptest.NewLogger(t)

			l := New()
			got, errD := l.ParsePayload(context.Background(), testLogger, tt.args.payload)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Listener.ParsePayload() got = %v, want %v", got, tt.want)
			}

			if errD == nil {
				if tt.wantErr != "" {
					t.Errorf("Listener.ParsePayload() err = nil, want %v", tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(errD.Detail, tt.wantErr) {
				t.Errorf("Listener.ParsePayload() err = %v, want %v", errD.Detail, tt.wantErr)
			}
		})
	}
}
