package config

import "context"

type ServerConfiguration struct {
	Notifications  []Notification `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	TraceHeaderKey string         `json:"traceHeaderKey,omitempty" yaml:"traceHeaderKey,omitempty"`
	//X-Cloud-Trace-Context
}

type Notification struct {
	Name                string            `json:"name,omitempty" yaml:"name,omitempty"`
	CelExpressionFilter string            `json:"celExpressionFilter,omitempty" yaml:"celExpressionFilter,omitempty"`
	Disabled            bool              `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	Type                string            `json:"type,omitempty" yaml:"type,omitempty"`
	Properties          map[string]string `json:"properties,omitempty" yaml:"properties,omitempty"`
	// Secrets             []Secret          `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

// type Secret struct {
// 	Name       string `json:"name,omitempty" yaml:"name,omitempty"`
// 	Type       string `json:"type,omitempty" yaml:"type,omitempty"`
// 	How should I di this??????
// 	Project    string `json:"project,omitempty" yaml:"project,omitempty"`
// 	SecretName string `json:"secretName,omitempty" yaml:"secretName,omitempty"`
// 	Version    string `json:"version,omitempty" yaml:"version,omitempty"`
// }

func NewServerConfiguration() *ServerConfiguration {
	return &ServerConfiguration{}
}

var config *ServerConfiguration

type ctxConfigKey struct{}

func FromCtx(ctx context.Context) *ServerConfiguration {
	if l, ok := ctx.Value(ctxConfigKey{}).(*ServerConfiguration); ok {
		return l
	} else if l := config; l != nil {
		return l
	}
	return NewServerConfiguration()
}

func WithCtx(ctx context.Context, l *ServerConfiguration) context.Context {
	if lp, ok := ctx.Value(ctxConfigKey{}).(*ServerConfiguration); ok {
		if lp == l {
			return ctx
		}
	}
	return context.WithValue(ctx, ctxConfigKey{}, l)
}
