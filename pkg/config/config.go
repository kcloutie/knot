package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kcloutie/knot/pkg/gcp"
	"github.com/kcloutie/knot/pkg/message"
	"go.uber.org/zap"
)

func AsBoolPointer(val bool) *bool {
	return &val
}

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
	Value        *string              `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom    *PropertyValueSource `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
	PayloadValue *PayloadValueRef     `json:"payloadValue,omitempty" yaml:"payloadValue,omitempty"`
	FromFile     *string              `json:"fromFile,omitempty" yaml:"fromFile,omitempty"`
	FromEnv      *string              `json:"fromEnv,omitempty" yaml:"fromEnv,omitempty"`
}

type NotificationProperty struct {
	Name        string              `json:"name" yaml:"name"`
	Required    *bool               `json:"required" yaml:"required"`
	Description string              `json:"description" yaml:"description"`
	Validation  *PropertyValidation `json:"validation" yaml:"validation"`
}

type PropertyValidation struct {
	ValidationRegex        string `json:"validationRegex,omitempty" yaml:"validationRegex,omitempty"`
	ValidationRegexMessage string `json:"validationRegexMessage,omitempty" yaml:"validationRegexMessage,omitempty"`
	AllowNullOrEmpty       *bool  `json:"allowNullOrEmpty,omitempty" yaml:"allowNullOrEmpty,omitempty"`
	MinLength              *int   `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength              *int   `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
}

func (pv *PropertyAndValue) GetValueProp() string {
	if pv.Value != nil {
		return *pv.Value
	}
	return ""
}

func (pv *PropertyAndValue) GetFromFileProp() string {
	if pv.FromFile != nil {
		return *pv.FromFile
	}
	return ""
}

func (pv *PropertyAndValue) GetFromEnvProp() string {
	if pv.FromEnv != nil {
		return *pv.FromEnv
	}
	return ""
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

func (o PropertyAndValue) GetValue(ctx context.Context, log *zap.Logger, data *message.NotificationData) (string, error) {
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

	if o.GetFromFileProp() != "" {
		fileBytes, err := os.ReadFile(o.GetFromFileProp())
		if err != nil {
			return "", fmt.Errorf("error reading file %s: %w", o.GetFromFileProp(), err)
		}
		return string(fileBytes), nil
	}

	if o.GetFromEnvProp() != "" {
		val := os.Getenv(o.GetFromEnvProp())
		if val == "" {
			log.Warn(fmt.Sprintf("environment variable %s is empty", o.GetFromEnvProp()))
		}
		log.Debug(fmt.Sprintf("environment variable %s value is %s", o.GetFromEnvProp(), val))
		return os.Getenv(o.GetFromEnvProp()), nil
	}

	return o.GetValueProp(), nil
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
