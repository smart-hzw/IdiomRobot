package websocket

import (
	"IdiomRobot/dto"
)

type uri string

const (
	messagesURI uri = "/channels/{channel_id}/messages"
)

var eventParseFuncMap = map[dto.OPCode]map[dto.EventType]eventParseFunc{
	dto.WSDispatchEvent: {
		dto.EventAtMessageCreate: atMessageHandler,
	},
}

type eventParseFunc func(event *dto.PayloadCommon, message []byte) error

func ParseAndHandle(payload *dto.PayloadCommon) error {
	// 指定类型的 handler
	if h, ok := eventParseFuncMap[dto.OPCode(payload.Op)][payload.T]; ok {
		return h(payload, payload.RawMessage)
	}

	return nil
}

func atMessageHandler(payload *dto.PayloadCommon, message []byte) error {
	data := &dto.ATMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.ATMessage != nil {
		return DefaultHandlers.ATMessage(payload, data)
	}
	return nil
}
