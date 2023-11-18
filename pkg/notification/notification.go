package notification

import (
	"context"
	"fmt"

	"github.com/kcloutie/knot/pkg/config"
	"github.com/kcloutie/knot/pkg/logger"
	"github.com/kcloutie/knot/pkg/message"
)

func Process(ctx context.Context, notification config.Notification, data *message.NotificationData) error {
	log := logger.FromCtx(ctx)
	log.Info(fmt.Sprintf("message '%s' processed using notification '%s'", data.ID, notification.Name))
	return nil
}
