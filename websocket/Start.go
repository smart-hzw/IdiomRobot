package websocket

import (
	"IdiomRobot/dto"
	"fmt"
	"log"
	"time"
)

var Pre = "QQBot "

// 需要一个绘画链，调用网关/gateway返回的数值
type WssInfo struct {
	URL               string            `json:"url"`    //wss链接
	Shards            uint32            `json:"shards"` //分片
	SessionStartLimit SessionStartLimit `json:"session_start_limit"`
}

// concurrencyTimeWindowSec 并发时间窗口，单位秒

// SessionStartLimit 链接频控信息
type SessionStartLimit struct {
	Total          uint32 `json:"total"`
	Remaining      uint32 `json:"remaining"`
	ResetAfter     uint32 `json:"reset_after"`
	MaxConcurrency uint32 `json:"max_concurrency"`
}

type Session struct {
	ID      string      `json:"id"`
	URL     string      `json:"url"`
	Token   Token       `json:"accessToken"`
	Intent  dto.Intent  `json:"intent"`
	LastSeq uint32      `json:"last_seq"`
	Shards  ShardConfig `json:"shards"`
}

type ShardConfig struct {
	ShardID    uint32
	ShardCount uint32
}

type Type string

// TokenType
const (
	TypeBot    Type = "Bot"
	TypeNormal Type = "Bearer"
)

type Token struct {
	AppID       uint64 `json:"appId"`
	AccessToken string `json:"access_token"`
	expires_in  string `json:"expires_in"`
	Type        Type
}

func (t *Token) GetString() string {
	if t.Type == TypeNormal {
		return t.AccessToken
	}
	return fmt.Sprintf("%v.%s", t.AppID, t.AccessToken)
}

type LinkChanManager struct {
	sessionChan chan Session
}

func New() *LinkChanManager {
	return &LinkChanManager{}
}

// Start 启动本地 session manager
func (linkChan *LinkChanManager) SessionChanStart(wssInfo *WssInfo, token *Token, intents *dto.Intent) error {
	log.Print("Info=====================开始启动session manager")
	//检查分片情况
	if err := CheckSessionLimit(wssInfo); err != nil {
		//如果分片受限，则返回
		log.Printf("Error=====================分片数量受限 %v", wssInfo)
		return err
	}
	startInterval := CalcInterval(wssInfo.SessionStartLimit.MaxConcurrency)

	//制作Session链
	linkChan.sessionChan = make(chan Session, wssInfo.Shards)
	//加入链
	for i := uint32(0); i < wssInfo.Shards; i++ {
		session := Session{
			URL:     wssInfo.URL,
			Token:   *token,
			Intent:  *intents,
			LastSeq: 0,
			Shards: ShardConfig{
				ShardID:    i,
				ShardCount: wssInfo.Shards,
			},
		}
		linkChan.sessionChan <- session
	}

	for session := range linkChan.sessionChan {
		// MaxConcurrency 代表的是每 5s 可以连多少个请求
		time.Sleep(startInterval)
		//遍历会话链，建立新连接
		go linkChan.CreateNewConnect(session)
	}
	return nil
}

// newConnect 启动一个新的连接，如果连接在监听过程中报错了，或者被远端关闭了链接，需要识别关闭的原因，能否继续 resume
// 如果能够 resume，则往 sessionChan 中放入带有 sessionID 的 session
// 如果不能，则清理掉 sessionID，将 session 放入 sessionChan 中
// session 的启动，交给 start 中的 for 循环执行，session 不自己递归进行重连，避免递归深度
func (linkChan *LinkChanManager) CreateNewConnect(session Session) {
	//记录当下
	defer func() {
		if err := recover(); err != nil {
			linkChan.sessionChan <- session
		}
	}()

	//创建实例
	wsClient := ClientImpl.Create(session)
	if err := wsClient.LinkWss(); err != nil {
		log.Printf("Error=====================wss链接失败 %v", err)
		linkChan.sessionChan <- session // 连接失败，丢回去队列排队重连
		return
	}
	log.Printf("Info=====================wss链接成功 %v")
	var err error
	// 如果 session id 不为空，则执行的是 resume 操作，如果为空，则执行的是 identify 操作
	if session.ID != "" {
		err = wsClient.Resume()
	} else {
		// 初次鉴权
		err = wsClient.Auth()
	}
	if err != nil {
		log.Printf("Error=====================鉴权或者重连失败 %v", err)
		return
	}
	log.Printf("Info=====================鉴权或重连成功 %v")
	if err := wsClient.Listening(); err != nil {
		log.Printf("Error=====================监听失败 %v", err)
		currentSession := wsClient.GetSession()
		linkChan.sessionChan <- *currentSession
		return
	}
}
