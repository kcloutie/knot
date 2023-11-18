package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/kcloutie/knot/pkg/http"
	"github.com/kcloutie/knot/pkg/message"
	"go.uber.org/zap"
)

type Listener struct {
	Name    string
	ApiPath string
}

func New() *Listener {
	return &Listener{
		Name:    "pub/sub",
		ApiPath: "pubsub",
	}
}

func (v *Listener) GetName() string {
	return v.Name
}

func (v *Listener) GetApiPath() string {
	return v.ApiPath
}

func (v *Listener) ParsePayload(ctx context.Context, log *zap.Logger, payload []byte) (*message.NotificationData, *http.ErrorDetail) {
	request := &pubsub.Message{}
	err := json.Unmarshal(payload, request)
	if err != nil {
		mess := fmt.Sprintf("Failed to unmarshal body to the pubsub.Message type. Error: %v", err)
		errD := &http.ErrorDetail{
			Type:     "unmarshal-body-data",
			Title:    "Unmarshal Body Data",
			Status:   400,
			Detail:   mess,
			Instance: v.GetApiPath(),
		}
		log.Error(mess)
		return nil, errD
	}

	notifyData, err := message.ToNotificationData(request)
	if err != nil {
		errD := &http.ErrorDetail{
			Type:     "convert-pubsub-message",
			Title:    "Convert Pub/Sub Message",
			Status:   400,
			Detail:   err.Error(),
			Instance: v.GetApiPath(),
		}
		log.Error(err.Error())
		return nil, errD
	}

	return &notifyData, nil
}
