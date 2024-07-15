package websocket

import (
	"IdiomRobot/dto"
	"log"
)

var DefaultHandlers struct {
	Ready       ReadyHandler
	ATMessage   ATMessageEventHandler
	Interaction InteractionEventHandler
}

func RegisterHandlers(handlers ...interface{}) dto.Intent {
	var i dto.Intent
	for _, h := range handlers {
		switch handle := h.(type) {
		case ReadyHandler:
			DefaultHandlers.Ready = handle
		case InteractionEventHandler:
			DefaultHandlers.Interaction = handle
			i = i | EventToIntent(dto.EventInteractionCreate)
		case ATMessageEventHandler:
			log.Printf("++++++++++++++++++++++++++++解析事件4444", handle)
			DefaultHandlers.ATMessage = handle
			i = i | EventToIntent(dto.EventAtMessageCreate)
		default:
		}
	}
	//i = i | registerHandlers(i, handlers...)
	return i

}

// Readyhandler事件处理
type ReadyHandler func(event *dto.PayloadCommon, data *dto.ReadyData)

// at事件处理
type ATMessageEventHandler func(event *dto.PayloadCommon, data *dto.ATMessageData) error

// 交互事件处理
type InteractionEventHandler func(event *dto.PayloadCommon, data *dto.WSInteractionData) error

var eventIntentMap = transposeIntentEventMap(dto.IntentEventMap)

func transposeIntentEventMap(input map[dto.Intent][]dto.EventType) map[dto.EventType]dto.Intent {
	result := make(map[dto.EventType]dto.Intent)
	for i, eventTypes := range input {
		for _, s := range eventTypes {
			result[s] = i
		}
	}
	return result
}

func EventToIntent(events ...dto.EventType) dto.Intent {
	var i dto.Intent
	for _, event := range events {
		i = i | eventIntentMap[event]
	}
	return i
}
