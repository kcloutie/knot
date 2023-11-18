package message

import (
	"reflect"
	"testing"
)

func TestNotificationData_AsCelData(t *testing.T) {

	tests := []struct {
		name string
		n    NotificationData
		want map[string]interface{}
	}{
		{
			name: "basic",
			n: NotificationData{
				Data: map[string]interface{}{
					"test": "123",
				},
				ID: "1",
				Attributes: map[string]string{
					"att1": "att1Val",
				},
			},
			want: map[string]interface{}{
				"data": map[string]interface{}{
					"test": "123",
				},
				"id": "1",
				"attributes": map[string]string{
					"att1": "att1Val",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.AsMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NotificationData.AsCelData() = %v, want %v", got, tt.want)
			}
		})
	}
}
