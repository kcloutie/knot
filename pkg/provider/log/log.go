package log

import (
	"context"
	"fmt"

	"github.com/kcloutie/knot/pkg/config"
	"github.com/kcloutie/knot/pkg/message"
	"github.com/kcloutie/knot/pkg/provider"
	"github.com/kcloutie/knot/pkg/template"
	"go.uber.org/zap"
)

var _ provider.ProviderInterface = (*Provider)(nil)

type Provider struct {
	log          *zap.Logger
	providerName string
	notification config.Notification
}

func New() *Provider {
	return &Provider{
		providerName: "log",
	}
}

func (v *Provider) SetLogger(logger *zap.Logger) {
	v.log = logger
}
func (v *Provider) GetName() string {
	return v.providerName
}

func (v *Provider) SetNotification(notification config.Notification) {
	v.notification = notification
}

func (v *Provider) SendNotification(ctx context.Context, data *message.NotificationData) error {
	_, err := provider.HasRequiredProperties(v.notification.Properties, v.GetRequiredPropertyNames())
	if err != nil {
		return err
	}
	message := v.notification.Properties["message"]

	templateConfig := template.NewRenderTemplateOptions()
	provider.SetGoTemplateOptionValues(v.log, &templateConfig, v.notification.Properties)

	renderedMessage, err := template.RenderTemplateValues(ctx, message, fmt.Sprintf("%s_%s", data.ID, v.providerName), data.AsMap(), []string{}, templateConfig)
	if err != nil {
		return err
	}

	v.log.Info(string(renderedMessage))
	return nil
}

func (v *Provider) GetHelp() string {
	return ""
}

func (v *Provider) GetRequiredPropertyNames() []string {
	return []string{"message"}
}
