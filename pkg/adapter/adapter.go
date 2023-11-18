package adapter

import (
	"github.com/kcloutie/knot/pkg/config"

	"github.com/kcloutie/knot/pkg/provider"
	logknot "github.com/kcloutie/knot/pkg/provider/log"
	"go.uber.org/zap"
)

func GetProviders() map[string]func(log *zap.Logger, notification config.Notification) provider.ProviderInterface {
	results := map[string]func(log *zap.Logger, notification config.Notification) provider.ProviderInterface{}

	results["log"] = func(log *zap.Logger, notification config.Notification) provider.ProviderInterface {
		pro := logknot.New()
		pro.SetLogger(log)
		pro.SetNotification(notification)
		return pro
	}

	return results
}
