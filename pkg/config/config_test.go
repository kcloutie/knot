package config

import (
	"context"
	"fmt"
	"hash/crc32"
	"reflect"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/kcloutie/knot/pkg/gcp"
	"github.com/kcloutie/knot/pkg/message"
	"go.uber.org/zap/zaptest"
)

func TestFromCtx(t *testing.T) {
	cfg := NewServerConfiguration()
	cfg.TraceHeaderKey = "HEADER"
	ctx := context.Background()
	ctx = WithCtx(ctx, cfg)
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want *ServerConfiguration
	}{
		{
			name: "already exists",
			args: args{
				ctx: ctx,
			},
			want: cfg,
		},
		{
			name: "new",
			args: args{
				ctx: context.Background(),
			},
			want: NewServerConfiguration(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromCtx(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	testVal := "test value"
	ignoreVal := "ignored"
	data := message.NotificationData{
		Data: map[string]interface{}{
			"test": "data test value",
			"child1": map[string]interface{}{
				"child2": map[string]interface{}{
					"test_child": "data test child value",
				},
			},
		},
		Attributes: map[string]string{},
		ID:         "test-id",
	}
	tests := []struct {
		name    string
		propVal PropertyAndValue
		data    *message.NotificationData
		want    string
		wantErr bool
	}{
		{
			name: "ValueFrom is nil",
			propVal: PropertyAndValue{
				Value: &testVal,
			},
			data:    &data,
			want:    "test value",
			wantErr: false,
		},
		{
			name: "ValueFrom is not nil but GcpSecretRef is nil",
			propVal: PropertyAndValue{
				Value: &testVal,
				ValueFrom: &PropertyValueSource{
					GcpSecretRef: nil,
				},
			},
			data:    &data,
			want:    "test value",
			wantErr: false,
		},
		{
			name: "ValueFrom and GcpSecretRef are not nil",
			propVal: PropertyAndValue{
				Value: &testVal,
				ValueFrom: &PropertyValueSource{
					GcpSecretRef: &GcpSecretRef{
						ProjectId: "test-project",
						Name:      "test-secret",
						Version:   "latest",
					},
				},
			},
			data: &data,
			want: "secret value",

			wantErr: false,
		},
		{
			name: "with PayloadValueRef",
			propVal: PropertyAndValue{
				Value: &ignoreVal,
				PayloadValue: &PayloadValueRef{
					PropertyPaths: []string{"data.test2", "data.test"},
				},
			},
			data:    &data,
			want:    "data test value",
			wantErr: false,
		},

		{
			name: "with PayloadValueRef with not existing path",
			propVal: PropertyAndValue{
				Value: &ignoreVal,
				PayloadValue: &PayloadValueRef{
					PropertyPaths: []string{"data.test2", "data.test3"},
				},
			},
			data:    &data,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Mock server for GCP Secret Manager API
			if tt.propVal.ValueFrom != nil && tt.propVal.ValueFrom.GcpSecretRef != nil {
				testServer, client := gcp.NewFakeServerAndClient(ctx, t)
				secretName := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", tt.propVal.ValueFrom.GcpSecretRef.ProjectId, tt.propVal.ValueFrom.GcpSecretRef.Name, tt.propVal.ValueFrom.GcpSecretRef.Version)
				crc32c := crc32.MakeTable(crc32.Castagnoli)
				checksum := int64(crc32.Checksum([]byte(tt.want), crc32c))

				testServer.Responses[secretName] = gcp.FakeSecretManagerServerResponse{
					Response: &secretmanagerpb.AccessSecretVersionResponse{
						Name: secretName,
						Payload: &secretmanagerpb.SecretPayload{
							Data:       []byte(tt.want),
							DataCrc32C: &checksum,
						},
					},
					Err: nil,
				}

				ctx = gcp.WithCtx(ctx, client)
				// gcp.SecretManagerBasePath = server.URL // Override the base path of the Secret Manager API
			}

			got, err := tt.propVal.GetValue(ctx, testLogger, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
