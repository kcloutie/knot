package api

import (
	"context"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/kcloutie/knot/pkg/adapter"
	"github.com/kcloutie/knot/pkg/config"
	"github.com/kcloutie/knot/pkg/http"
	"github.com/kcloutie/knot/pkg/listener"
	"github.com/kcloutie/knot/pkg/matcher"

	"go.uber.org/zap"
)

func ExecuteListener(ctx context.Context, c *gin.Context, listener listener.ListenerInterface) {
	cfg := config.FromCtx(ctx)
	var log *zap.Logger
	log, ctx = http.SetCommonLoggingAttributes(ctx, c)
	slog := log.Sugar()
	if c.Request.Body == nil {
		errorMes := "request body was empty, request cannot be processed"
		errD := &http.ErrorDetail{
			Type:     listener.GetName() + "-get-request-body",
			Title:    listener.GetName() + " Get Request Body",
			Status:   400,
			Detail:   errorMes,
			Instance: listener.GetApiPath(),
		}
		log.Error(errorMes)
		c.JSON(int(errD.Status), errD)
		return
	}
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		errD := &http.ErrorDetail{
			Type:     listener.GetName() + "-get-request-body",
			Title:    listener.GetName() + " Get Request Body",
			Status:   400,
			Detail:   err.Error(),
			Instance: listener.GetApiPath(),
		}
		log.Error(err.Error())
		c.JSON(int(errD.Status), errD)
		return
	}
	notifyData, errD := listener.ParsePayload(ctx, log, payload)
	if errD != nil {
		log.Error(errD.Detail)
		c.JSON(int(errD.Status), errD)
		return
	}

	providerFunctions := adapter.GetProviders()

	if len(cfg.Notifications) == 0 {
		slog.Warnf("no notifications configured for listener '%s'", listener.GetName())
	}

	for _, not := range cfg.Notifications {
		matches, err := matcher.Matches(ctx, not, notifyData)
		if err != nil {
			errD := &http.ErrorDetail{
				Type:     listener.GetName() + "-listener-message",
				Title:    listener.GetName() + "Match Message",
				Status:   400,
				Detail:   err.Error(),
				Instance: listener.GetApiPath(),
			}
			log.Error(err.Error())
			c.JSON(int(errD.Status), errD)
			return
		}
		if !matches {
			slog.Debugf("notification '%s' does not match message", not.Name)
			continue
		}

		proNewFunc, exists := providerFunctions[not.Type]
		if !exists {
			errD := &http.ErrorDetail{
				Type:     listener.GetName() + "-exists",
				Title:    listener.GetName() + " Exists",
				Status:   400,
				Detail:   fmt.Sprintf("notification type of '%s' does not exist. Check the notification type of the '%s' notification", not.Type, not.Name),
				Instance: listener.GetApiPath(),
			}
			log.Error(err.Error())
			c.JSON(int(errD.Status), errD)
			return
		}
		provider := proNewFunc(log, not)
		slog.Debugf("sending notification '%s' to provider '%s'", not.Name, provider.GetName())
		err = provider.SendNotification(ctx, notifyData)
		if err != nil {
			errD := &http.ErrorDetail{
				Type:     listener.GetName() + "-" + provider.GetName() + "-send-notification",
				Title:    listener.GetName() + "-" + provider.GetName() + " Send Notification",
				Status:   400,
				Detail:   err.Error(),
				Instance: listener.GetApiPath(),
			}
			log.Error(err.Error())
			c.JSON(int(errD.Status), errD)
			return
		}
		slog.Debugf("notification '%s' sent to provider '%s'", not.Name, provider.GetName())
	}
}
