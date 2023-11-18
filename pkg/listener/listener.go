package listener

import "github.com/kcloutie/knot/pkg/listener/pubsub"

func GetListeners() []ListenerInterface {
	listeners := []ListenerInterface{}
	listeners = append(listeners, pubsub.New())
	return listeners
}
