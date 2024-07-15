package main

import "log"

var DefaultHandlers struct {
	Ready       ReadyHandler
	ATMessage   ATMessageEventHandler
	Interaction InteractionEventHandler
}

func RegisterHandlers(handlers ...interface{}) Intent {
	var i Intent
	for _, h := range handlers {
		switch handle := h.(type) {
		case ReadyHandler:
			DefaultHandlers.Ready = handle
		case InteractionEventHandler:
			DefaultHandlers.Interaction = handle
			i = i | EventToIntent(EventInteractionCreate)
		case ATMessageEventHandler:
			log.Printf("++++++++++++++++++++++++++++解析事件4444", handle)
			DefaultHandlers.ATMessage = handle
			i = i | EventToIntent(EventAtMessageCreate)
		default:
		}
	}
	//i = i | registerHandlers(i, handlers...)
	return i

}

// Readyhandler事件处理
type ReadyHandler func(event *PayloadCommon, data *ReadyData)

// at事件处理
type ATMessageEventHandler func(event *PayloadCommon, data *ATMessageData) error

// 交互事件处理
type InteractionEventHandler func(event *PayloadCommon, data *WSInteractionData) error

var eventIntentMap = transposeIntentEventMap(intentEventMap)

func transposeIntentEventMap(input map[Intent][]EventType) map[EventType]Intent {
	result := make(map[EventType]Intent)
	for i, eventTypes := range input {
		for _, s := range eventTypes {
			result[s] = i
		}
	}
	return result
}

func EventToIntent(events ...EventType) Intent {
	var i Intent
	for _, event := range events {
		i = i | eventIntentMap[event]
	}
	return i
}
