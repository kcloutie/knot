package provider

import (
	"context"

	"github.com/kcloutie/knot/pkg/config"
	"github.com/kcloutie/knot/pkg/message"
	"go.uber.org/zap"
)

type ProviderInterface interface {
	GetName() string
	SetLogger(logger *zap.Logger)
	SetNotification(notification config.Notification)
	SendNotification(ctx context.Context, data *message.NotificationData) error
	GetHelp() string
	GetRequiredPropertyNames() []string
}
