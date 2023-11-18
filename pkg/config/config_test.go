package config

import (
	"context"
	"reflect"
	"testing"
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
