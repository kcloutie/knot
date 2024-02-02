package message

import (
	"reflect"
	"testing"

	"cloud.google.com/go/pubsub"
)

func TestToNotificationData(t *testing.T) {
	tests := []struct {
		name    string
		message pubsub.Message
		want    NotificationData
		wantErr bool
	}{
		{
			name: "All fields filled",
			message: pubsub.Message{
				Attributes: map[string]string{
					"att1": "att1Val",
				},
				Data: []byte(`{"test":"123"}`),
				ID:   "1",
			},
			want: NotificationData{
				Attributes: map[string]string{
					"att1": "att1Val",
				},
				Data: map[string]interface{}{
					"test": "123",
				},
				ID: "1",
			},
			wantErr: false,
		},
		{
			name: "Some fields empty",
			message: pubsub.Message{
				Attributes: map[string]string{},
				Data:       []byte(`{}`),
				ID:         "1",
			},
			want: NotificationData{
				Attributes: map[string]string{},
				Data:       map[string]interface{}{},
				ID:         "1",
			},
			wantErr: false,
		},
		{
			name: "Invalid JSON data",
			message: pubsub.Message{
				Attributes: map[string]string{
					"att1": "att1Val",
				},
				Data: []byte(`dude`),
				ID:   "1",
			},
			want: NotificationData{
				Attributes: map[string]string{
					"att1": "att1Val",
				},
				ID: "1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToNotificationData(&tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToNotificationData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToNotificationData() got = %v, want %v", got, tt.want)
			}
		})
	}
}
