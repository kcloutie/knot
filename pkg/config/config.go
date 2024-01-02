package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kcloutie/knot/pkg/gcp"
	"github.com/kcloutie/knot/pkg/message"
)

type ServerConfiguration struct {
	Notifications  []Notification `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	TraceHeaderKey string         `json:"traceHeaderKey,omitempty" yaml:"traceHeaderKey,omitempty"`
	//X-Cloud-Trace-Context
}

type Notification struct {
	Name                string                      `json:"name,omitempty" yaml:"name,omitempty"`
	CelExpressionFilter string                      `json:"celExpressionFilter,omitempty" yaml:"celExpressionFilter,omitempty"`
	Disabled            bool                        `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	Type                string                      `json:"type,omitempty" yaml:"type,omitempty"`
	Properties          map[string]PropertyAndValue `json:"properties,omitempty" yaml:"properties,omitempty"`
	// Secrets             []Secret          `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

type PropertyAndValue struct {
	// Name         string               `json:"name,omitempty" yaml:"name,omitempty"`
	Value        string               `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom    *PropertyValueSource `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
	PayloadValue *PayloadValueRef     `json:"payloadValue,omitempty" yaml:"payloadValue,omitempty"`
	FromFile     string               `json:"fromFile,omitempty" yaml:"fromFile,omitempty"`
}

type PropertyValueSource struct {
	GcpSecretRef *GcpSecretRef `json:"secretKeyRef,omitempty" yaml:"secretKeyRef,omitempty"`
}

type GcpSecretRef struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	ProjectId string `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	Version   string `json:"version,omitempty" yaml:"version,omitempty"`
}

type PayloadValueRef struct {
	PropertyPaths []string `json:"propertyPaths,omitempty" yaml:"propertyPaths,omitempty"`
}

func (o PropertyAndValue) GetValue(ctx context.Context, data *message.NotificationData) (string, error) {
	if o.ValueFrom != nil && o.ValueFrom.GcpSecretRef != nil {
		secClient := gcp.FromCtx(ctx)
		if secClient != nil {
			defer secClient.Close()
		}
		val, err := gcp.GetSecret(ctx, secClient, o.ValueFrom.GcpSecretRef.ProjectId, o.ValueFrom.GcpSecretRef.Name, o.ValueFrom.GcpSecretRef.Version)
		if err != nil {
			return "", fmt.Errorf("error getting secret %s/%s/%s: %w", o.ValueFrom.GcpSecretRef.ProjectId, o.ValueFrom.GcpSecretRef.Name, o.ValueFrom.GcpSecretRef.Version, err)
		}
		return val, nil
	}
	if o.PayloadValue != nil && len(o.PayloadValue.PropertyPaths) != 0 {
		errs := []string{}
		for _, path := range o.PayloadValue.PropertyPaths {
			val, err := data.GetPropertyValue(path)
			if err == nil {
				return val, nil
			} else {
				errs = append(errs, err.Error())
				continue
			}
		}
		return "", fmt.Errorf("error getting property value from the following paths '%s'. Errors: %s", strings.Join(o.PayloadValue.PropertyPaths, ", "), strings.Join(errs, ", "))
	}

	if o.FromFile != "" {
		fileBytes, err := os.ReadFile(o.FromFile)
		if err != nil {
			return "", fmt.Errorf("error reading file %s: %w", o.FromFile, err)
		}
		return string(fileBytes), nil
	}

	return o.Value, nil
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
