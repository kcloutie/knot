package api

// import (
// 	"context"

// 	"github.com/gin-gonic/gin"
// 	"github.com/kcloutie/knot/pkg/config"
// 	"github.com/kcloutie/knot/pkg/http"
// 	knotpubsub "github.com/kcloutie/knot/pkg/listener/pubsub"
// 	"github.com/kcloutie/knot/pkg/matcher"
// 	"github.com/kcloutie/knot/pkg/notification"
// 	"go.uber.org/zap"
// )

// func PubSub(ctx context.Context, c *gin.Context) {
// 	cfg := config.FromCtx(ctx)
// 	var log *zap.Logger

// 	log, ctx = http.SetCommonLoggingAttributes(ctx, c)

// 	listener := knotpubsub.New()
// 	notifyData, errD := listener.ParsePayload(ctx, c)
// 	if errD != nil {
// 		log.Error(errD.Detail)
// 		c.JSON(int(errD.Status), errD)
// 		return
// 	}
// 	for _, not := range cfg.Notifications {
// 		matches, err := matcher.Matches(ctx, not, notifyData)
// 		if err != nil {
// 			errD := &http.ErrorDetail{
// 				Type:     "match-pubsub-message",
// 				Title:    "Match Pub/Sub Message",
// 				Status:   400,
// 				Detail:   err.Error(),
// 				Instance: listener.ApiPath,
// 			}
// 			log.Error(err.Error())
// 			c.JSON(int(errD.Status), errD)
// 			return
// 		}
// 		if !matches {

// 			continue
// 		}
// 		notification.Process(ctx, not, notifyData)
// 	}

// }
