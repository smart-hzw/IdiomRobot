package dto

const (
	EventAtMessageCreate   EventType = "AT_MESSAGE_CREATE"
	EventInteractionCreate EventType = "INTERACTION_CREATE"
)

var IntentEventMap = map[Intent][]EventType{
	IntentGuildAtMessage: {EventAtMessageCreate},
	IntentInteraction:    {EventInteractionCreate},
}
