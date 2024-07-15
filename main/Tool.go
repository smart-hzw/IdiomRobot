package main

type uri string

const (
	messagesURI uri = "/channels/{channel_id}/messages"
)

var eventParseFuncMap = map[OPCode]map[EventType]eventParseFunc{
	WSDispatchEvent: {
		EventAtMessageCreate: atMessageHandler,
	},
}

type eventParseFunc func(event *PayloadCommon, message []byte) error

func ParseAndHandle(payload *PayloadCommon) error {
	// 指定类型的 handler
	if h, ok := eventParseFuncMap[OPCode(payload.Op)][payload.T]; ok {
		return h(payload, payload.RawMessage)
	}

	return nil
}

func atMessageHandler(payload *PayloadCommon, message []byte) error {
	data := &ATMessageData{}
	if err := ParseData(message, data); err != nil {
		return err
	}
	if DefaultHandlers.ATMessage != nil {
		return DefaultHandlers.ATMessage(payload, data)
	}
	return nil
}
