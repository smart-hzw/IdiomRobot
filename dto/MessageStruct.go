package dto

import (
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Avatar           string `json:"avatar"`
	Bot              bool   `json:"bot"`
	UnionOpenID      string `json:"union_openid"`       // 特殊关联应用的 openid
	UnionUserAccount string `json:"union_user_account"` // 机器人关联的用户信息，与union_openid关联的应用是同一个
}

type Timestamp string

// Time 时间字符串格式转换
func (t Timestamp) Time() (time.Time, error) {
	return time.Parse(time.RFC3339, string(t))
}

// Message 消息结构体定义
type Message struct {
	// 消息ID
	ID string `json:"id"`
	// 子频道ID
	ChannelID string `json:"channel_id"`
	// 频道ID
	GuildID string `json:"guild_id"`
	// 内容
	Content string `json:"content"`
	// 发送时间
	Timestamp Timestamp `json:"timestamp"`
	// 消息编辑时间
	EditedTimestamp Timestamp `json:"edited_timestamp"`
	// 是否@all
	MentionEveryone bool `json:"mention_everyone"`
	// 消息发送方
	Author *User `json:"author"`
	// 消息发送方Author的member属性，只是部分属性
	Member *Member `json:"member"`
	// 附件
	Attachments []*MessageAttachment `json:"attachments"`
	// 结构化消息-embeds
	Embeds []*Embed `json:"embeds"`
	// 消息中的提醒信息(@)列表
	Mentions []*User `json:"mentions"`
	// ark 消息
	Ark *Ark `json:"ark"`
	// 私信消息
	DirectMessage bool `json:"direct_message"`
	// 子频道 seq，用于消息间的排序，seq 在同一子频道中按从先到后的顺序递增，不同的子频道之前消息无法排序
	SeqInChannel string `json:"seq_in_channel"`
	// 引用的消息
	MessageReference *MessageReference `json:"message_reference,omitempty"`
	// 私信场景下，该字段用来标识从哪个频道发起的私信
	SrcGuildID string `json:"src_guild_id"`
}

type MessageReference struct {
	MessageID             string `json:"message_id"`               // 消息 id
	IgnoreGetMessageError bool   `json:"ignore_get_message_error"` // 是否忽律获取消息失败错误
}

type Member struct {
	GuildID  string    `json:"guild_id"`
	JoinedAt Timestamp `json:"joined_at"`
	Nick     string    `json:"nick"`
	User     *User     `json:"user"`
	Roles    []string  `json:"roles"`
	OpUserID string    `json:"op_user_id,omitempty"`
}

// DeleteHistoryMsgDay 消息撤回天数
type DeleteHistoryMsgDay = int

// MemberDeleteOpts 删除成员额外参数
type MemberDeleteOpts struct {
	AddBlackList         bool                `json:"add_blacklist"`
	DeleteHistoryMsgDays DeleteHistoryMsgDay `json:"delete_history_msg_days"`
}

type Ark struct {
	TemplateID int      `json:"template_id,omitempty"` // ark 模版 ID
	KV         []*ArkKV `json:"kv,omitempty"`          // ArkKV 数组
}

// ArkKV Ark 键值对
type ArkKV struct {
	Key   string    `json:"key,omitempty"`
	Value string    `json:"value,omitempty"`
	Obj   []*ArkObj `json:"obj,omitempty"`
}

type ArkObj struct {
	ObjKV []*ArkObjKV `json:"obj_kv,omitempty"`
}

// ArkObjKV Ark 对象键值对
type ArkObjKV struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type Interaction struct {
	ID            string           `json:"id,omitempty"`             // 互动行为唯一标识
	ApplicationID string           `json:"application_id,omitempty"` // 应用ID
	Type          InteractionType  `json:"type,omitempty"`           // 互动类型
	Data          *InteractionData `json:"data,omitempty"`           // 互动数据
	GuildID       string           `json:"guild_id,omitempty"`       // 频道 ID
	ChannelID     string           `json:"channel_id,omitempty"`     // 子频道 ID
	Version       uint32           `json:"version,omitempty"`        //	版本，默认为 1
}

type InteractionData struct {
	Name     string              `json:"name,omitempty"`     // 标题
	Type     InteractionDataType `json:"type,omitempty"`     //	数据类型，不同数据类型对应不同的 resolved 数据
	Resolved json.RawMessage     `json:"resolved,omitempty"` // 跟不同的互动类型和数据类型有关系的数据
}

type MessageToCreate struct {
	Content string `json:"content,omitempty"`
	Embed   *Embed `json:"embed,omitempty"`
	Ark     *Ark   `json:"ark,omitempty"`
	Image   string `json:"image,omitempty"`
	// 要回复的消息id，为空是主动消息，公域机器人会异步审核，不为空是被动消息，公域机器人会校验语料
	MsgID            string            `json:"msg_id,omitempty"`
	MessageReference *MessageReference `json:"message_reference,omitempty"`
	Markdown         *Markdown         `json:"markdown,omitempty"`
	Keyboard         *MessageKeyboard  `json:"keyboard,omitempty"` // 消息按钮组件
	EventID          string            `json:"event_id,omitempty"` // 要回复的事件id, 逻辑同MsgID
}

// Embed 结构
type Embed struct {
	Title       string                `json:"title,omitempty"`
	Description string                `json:"description,omitempty"`
	Prompt      string                `json:"prompt"` // 消息弹窗内容，消息列表摘要
	Thumbnail   MessageEmbedThumbnail `json:"thumbnail,omitempty"`
	Fields      []*EmbedField         `json:"fields,omitempty"`
}

// MessageEmbedThumbnail embed 消息的缩略图对象
type MessageEmbedThumbnail struct {
	URL string `json:"url"`
}

// EmbedField Embed字段描述
type EmbedField struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// MessageAttachment 附件定义
type MessageAttachment struct {
	URL string `json:"url"`
}

// MessageReactionUsers 消息表情表态用户列表
type MessageReactionUsers struct {
	Users  []*User `json:"users,omitempty"`
	Cookie string  `json:"cookie,omitempty"`
	IsEnd  bool    `json:"is_end,omitempty"`
}

// Markdown markdown 消息
type Markdown struct {
	TemplateID int               `json:"template_id"` // 模版 id
	Params     []*MarkdownParams `json:"params"`      // 模版参数
	Content    string            `json:"content"`     // 原生 markdown
}

// MarkdownParams markdown 模版参数 键值对
type MarkdownParams struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

// InteractionDataType 互动数据类型
type InteractionDataType uint32

// InteractionType 互动类型
type InteractionType uint32

type ATMessageData Message

type WSInteractionData Interaction

func Emoji(id int) string {
	return fmt.Sprintf("<emoji:%d>", id)
}
