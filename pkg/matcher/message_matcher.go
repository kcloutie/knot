package matcher

import (
	"context"

	"github.com/google/cel-go/common/types"
	"github.com/kcloutie/knot/pkg/cel"
	"github.com/kcloutie/knot/pkg/config"
	"github.com/kcloutie/knot/pkg/logger"
	"github.com/kcloutie/knot/pkg/message"
	"go.uber.org/zap"
)

func Matches(ctx context.Context, notification config.Notification, data *message.NotificationData) (bool, error) {
	log := logger.FromCtx(ctx).With(zap.String("notificationName", notification.Name))
	if notification.Disabled {
		log.Debug("message did not match notification because notification was disabled")
		return false, nil

	}
	if notification.CelExpressionFilter == "" {
		log.Debug("message matched notification! Notification does not contain any CEL filtering so it matches everything")
		return true, nil
	}

	matches, err := cel.CelEvaluate(ctx, notification.CelExpressionFilter, message.GetCelDecl(), data.AsMap())
	if err != nil {
		return false, err
	}
	if matches == types.True {
		log.Debug("message matched notification CEL filtering")
		return true, nil
	}
	log.Debug("message did not match notification CEL filtering")
	return false, nil

}
