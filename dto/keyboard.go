package dto

// ActionType 按钮操作类型
type ActionType uint32

// PermissionType 按钮的权限类型
type PermissionType uint32

// MessageKeyboard 消息按钮组件
type MessageKeyboard struct {
	ID      string          `json:"id,omitempty"`      // 消息按钮组件模板 ID
	Content *CustomKeyboard `json:"content,omitempty"` // 消息按钮组件自定义内容
}

// CustomKeyboard 自定义 Keyboard
type CustomKeyboard struct {
	Rows []*Row `json:"rows,omitempty"` // 行数组
}

// Row 每行结构
type Row struct {
	Buttons []*Button `json:"buttons,omitempty"` // 每行按钮
}

// Button 单个按纽
type Button struct {
	ID         string      `json:"id,omitempty"`          // 按钮 ID
	RenderData *RenderData `json:"render_data,omitempty"` // 渲染展示字段
	Action     *Action     `json:"action,omitempty"`      // 该按纽操作相关字段
}

// RenderData  按纽渲染展示
type RenderData struct {
	Label        string `json:"label,omitempty"`         // 按纽上的文字
	VisitedLabel string `json:"visited_label,omitempty"` // 点击后按纽上文字
	Style        int    `json:"style,omitempty"`         // 按钮样式，0：灰色线框，1：蓝色线框
}

// Action 按纽点击操作
type Action struct {
	Type                 ActionType  `json:"type,omitempty"`                     // 操作类型
	Permission           *Permission `son:"permission,omitempty"`                // 可操作
	ClickLimit           uint32      `json:"click_limit,omitempty"`              // 可点击的次数, 默认不限
	Data                 string      `json:"data,omitempty"`                     // 操作相关数据
	AtBotShowChannelList bool        `json:"at_bot_show_channel_list,omitempty"` // false:当前 true:弹出展示子频道选择器
}

// Permission 按纽操作权限
type Permission struct {
	// Type 操作权限类型
	Type PermissionType `json:"type,omitempty"`
	// SpecifyRoleIDs 身份组
	SpecifyRoleIDs []string `json:"specify_role_ids,omitempty"`
	// SpecifyUserIDs 指定 UserID
	SpecifyUserIDs []string `json:"specify_user_ids,omitempty"`
}
