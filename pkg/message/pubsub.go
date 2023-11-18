package message

import (
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
)

func ToNotificationData(message *pubsub.Message) (NotificationData, error) {
	results := NotificationData{
		Attributes: message.Attributes,
		ID:         message.ID,
	}
	data := map[string]interface{}{}
	err := json.Unmarshal(message.Data, &data)
	if err != nil {
		return results, fmt.Errorf("failed to unmarshal the pub/sub data property of the message - %v", err)
	}
	results.Data = data
	return results, nil
}
