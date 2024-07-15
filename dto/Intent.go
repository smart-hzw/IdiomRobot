package dto

// Intent 类型
type Intent int

const (
	IntentGuilds Intent = 1 << iota

	IntentInteraction Intent = 1 << 26 // 互动事件

	IntentGuildAtMessage Intent = 1 << 30 // 只接收@消息事件

)
