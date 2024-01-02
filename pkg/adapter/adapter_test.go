package adapter

import (
	"testing"

	"github.com/kcloutie/knot/pkg/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestGetProviders(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	notification := config.Notification{}

	tests := []struct {
		name         string
		providerType string
		wantExists   bool
	}{
		{
			name:         "Provider type is log",
			providerType: "log",
			wantExists:   true,
		},
		{
			name:         "Provider type is not log",
			providerType: "not_log",
			wantExists:   false,
		},
		{
			name:         "Provider type is github/comment",
			providerType: "github/comment",
			wantExists:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerFunctions := GetProviders()
			pFunc, exists := providerFunctions[tt.providerType]
			assert.Equal(t, tt.wantExists, exists)
			if !exists {
				return
			}
			provider := pFunc(testLogger, notification)
			assert.NotNil(t, provider)
			assert.Equal(t, tt.providerType, provider.GetName())
		})
	}
}
