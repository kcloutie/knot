package listener

import (
	"context"

	"github.com/kcloutie/knot/pkg/http"
	"github.com/kcloutie/knot/pkg/message"
	"go.uber.org/zap"
)

type ListenerInterface interface {
	GetName() string
	GetApiPath() string
	ParsePayload(ctx context.Context, log *zap.Logger, payload []byte) (*message.NotificationData, *http.ErrorDetail)
}
