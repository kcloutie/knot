package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kcloutie/knot/pkg/params/settings"
	"github.com/stretchr/testify/assert"
)

func TestPubSubListener(t *testing.T) {
	ctx := context.Background()
	router := CreateRouter(ctx, 1)

	badBody := strings.NewReader(`{"test":"test"}`)

	type args struct {
		method string
		url    string
		body   io.Reader
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
		wantBody string
	}{
		{
			name: "no body",
			args: args{
				method: "POST",
				url:    "/api/v1/" + settings.PubSubEndpoint,
				body:   nil,
			},
			wantCode: 400,
			wantBody: "{\"type\":\"pub/sub-get-request-body\",\"title\":\"pub/sub Get Request Body\",\"status\":400,\"detail\":\"request body was empty, request cannot be processed\",\"instance\":\"pubsub\"}",
		},

		{
			name: "bad body",
			args: args{
				method: "POST",
				url:    "/api/v1/" + settings.PubSubEndpoint,
				body:   badBody,
			},
			wantCode: 400,
			wantBody: "{\"type\":\"convert-pubsub-message\",\"title\":\"Convert Pub/Sub Message\",\"status\":400,\"detail\":\"failed to unmarshal the pub/sub data property of the message - unexpected end of JSON input\",\"instance\":\"pubsub\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.args.method, tt.args.url, tt.args.body)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}
