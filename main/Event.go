package main

const (
	EventAtMessageCreate   EventType = "AT_MESSAGE_CREATE"
	EventInteractionCreate EventType = "INTERACTION_CREATE"
)

var intentEventMap = map[Intent][]EventType{
	IntentGuildAtMessage: {EventAtMessageCreate},
	IntentInteraction:    {EventInteractionCreate},
}
