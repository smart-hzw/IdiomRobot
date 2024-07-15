package dto

type EventType string
type PayLoadBase struct {
	Op int       `json:"op"`
	S  int       `json:"s,omitempty"`
	T  EventType `json:"t,omitempty"`
}
type PayloadCommon struct {
	PayLoadBase
	CommonData interface{} `json:"d"`
	RawMessage []byte      `json:"-"`
}

type IdentityData struct {
	Token      string   `json:"token"`
	Intents    Intent   `json:"intents"`
	Shard      []uint32 `json:"shard"` // array of two integers (shard_id, num_shards)
	Properties struct {
		Os      string `json:"$os,omitempty"`
		Browser string `json:"$browser,omitempty"`
		Device  string `json:"$device,omitempty"`
	} `json:"properties,omitempty"`
}

type ResumeData struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       uint32 `json:"seq"`
}

type HelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type ReadyData struct {
	Version   int    `json:"version"`
	SessionID string `json:"session_id"`
	User      struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Bot      bool   `json:"bot"`
	} `json:"user"`
	Shard []uint32 `json:"shard"`
}

type WSUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Bot      bool   `json:"bot"`
}
