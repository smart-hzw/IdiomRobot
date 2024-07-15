package main

import (
	"encoding/json"
	"fmt"
	wss "github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"log"
	"syscall"
	"time"
)

const DefaultQueueSize = 10000

var (
	ClientImpl   Websocket
	ResumeSignal syscall.Signal
)

func Register(ws Websocket) {
	ClientImpl = ws
}

func Setup() {
	client := &Client{}
	Register(client)
}

// 每一个websocket都会执行的
type Websocket interface {
	//新建websocket
	Create(session Session) Websocket
	LinkWss() error
	Auth() error
	Listening() error
	//关闭链接
	Close() error
	//wss发送消息
	WriteMessage(message *PayloadCommon) error
	//重连
	Resume() error
	//获取Session信息
	GetSession() *Session
	ReadMessage()
}

//实现以上接口

//var ClientImpl Websocket

// 要有一个结构体来存储WebSocket对象的数据
type MessageChan chan *PayloadCommon
type closeErrorChan chan error
type Client struct {
	version         int
	Conn            *wss.Conn   //负责链接
	Session         *Session    //存储Session会话信息
	messageQueue    MessageChan //存储消息链表
	closeChan       closeErrorChan
	user            *WSUser
	heartBeatTicker *time.Ticker //维持心跳
}

func (c *Client) ReadMessage() {
	messageType, p, err := c.Conn.ReadMessage()
	if err != nil {
		log.Printf("Error===================消息读取失败", err)
	}
	log.Printf("Info===================%V", string(p), messageType)
}

func (c *Client) Create(session Session) Websocket {
	cc := &Client{
		Session:         &session,
		messageQueue:    make(MessageChan, DefaultQueueSize),
		closeChan:       make(closeErrorChan, 10),
		heartBeatTicker: time.NewTicker(60 * time.Second),
	}
	return cc
}

func (c *Client) LinkWss() error {
	if c.Session.URL == "" {
		return nil
	}
	var err error
	c.Conn, _, err = wss.DefaultDialer.Dial(c.Session.URL, nil)
	if err != nil {
		log.Printf("Error=====================wss连接失败 %v", err)
	}
	log.Printf("Info=====================连接成功", c.Session)
	return nil
}

func (c *Client) Auth() error {
	if c.Session.Intent == 0 {
		c.Session.Intent = IntentGuilds
	}
	data := &IdentityData{
		Token:   pre + c.Session.Token.AccessToken,
		Intents: c.Session.Intent,
		Shard: []uint32{
			c.Session.Shards.ShardID,
			c.Session.Shards.ShardCount,
			//0,
			//4,
		},
	}
	payload := &PayloadCommon{
		CommonData: *data,
	}

	payload.Op = 2
	log.Printf("Info========================查看鉴权请求体%v", payload)
	return c.WriteMessage(payload)
}

func (c *Client) Listening() error {
	defer c.Conn.Close()
	//读取消息队列
	go c.readMessageToQueue()
	//处理消息
	go c.listenMessageAndHandle()

	ticker := time.NewTicker(time.Minute) // 每5秒发送一次心跳
	defer ticker.Stop()
	//定时发送心跳
	for {
		select {
		case <-ticker.C:
			//发送心跳
			log.Printf("Info=====================心跳检测开始")
			heartBeatEvent := &PayloadCommon{
				PayLoadBase: PayLoadBase{
					Op: 1,
				},
				CommonData: c.Session.LastSeq,
			}
			err := c.WriteMessage(heartBeatEvent)
			//c.ReadMessage()
			if err != nil {
				log.Println("Error=====================心跳检测失败", err)
			}
		}

	}
	return nil
}

func (c *Client) Close() error {
	err := c.Conn.Close()
	if err != nil {
		log.Printf("Error=====================wss连接关闭失败 %v", err)
	}
	//c.heartBeatTicker.Stop()
	return nil
}

func (c *Client) WriteMessage(message *PayloadCommon) error {
	messageJson, _ := json.Marshal(message)
	log.Printf("Info=====================%v 消息正在发送中", string(messageJson))
	err := c.Conn.WriteMessage(wss.TextMessage, messageJson)
	if err != nil {
		log.Printf("Error=====================%v 消息发送失败")
		//写入关闭链
		return err
	}
	return nil
}

func (c *Client) Resume() error {
	data := &ResumeData{
		Token:     c.Session.Token.AccessToken,
		SessionID: c.Session.ID,
		Seq:       c.Session.LastSeq,
	}
	payload := &PayloadCommon{
		CommonData: *data,
	}
	payload.Op = 6
	return c.WriteMessage(payload)
}

func (c *Client) GetSession() *Session {
	return c.Session
}

// 将消息读取到消息队列中
func (c *Client) readMessageToQueue() {
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error=====================%v 消息队列的消息读取失败", err)
			close(c.messageQueue)
			c.closeChan <- err
			return
		}
		payload := &PayloadCommon{}
		err = json.Unmarshal(message, payload)
		if err != nil {
			log.Printf("Error=====================json解析失败 %v", err)
			continue
		}
		payload.RawMessage = message
		log.Printf("Info=====================消息数据处理中", c.Session, OPMeans(OPCode(payload.Op)), string(message))
		if c.isHandleBuildIn(payload) {
			continue
		}
		c.messageQueue <- payload
	}
}

func (c *Client) listenMessageAndHandle() {
	defer func() {
		if err := recover(); err != nil {
			//PanicHandler(err, c.session)
			c.closeChan <- fmt.Errorf("panic: %v", err)
		}
	}()
	for payload := range c.messageQueue {
		if payload.S > 0 {
			c.Session.LastSeq = uint32(payload.S)
		}
		// ready 事件需要特殊处理
		if payload.T == "READY" {
			c.readyHandler(payload)
			continue
		}
		if err := ParseAndHandle(payload); err != nil {
			log.Println("Info====================解析事件失败！", err)
		}
	}
	log.Printf("Info=====================消息队列关闭", c.Session)
}

func (c *Client) readyHandler(payload *PayloadCommon) {
	readyData := &ReadyData{}
	if err := ParseData(payload.RawMessage, readyData); err != nil {
		log.Printf("Error=====================Redy数据转换失败 %v")
	}
	c.version = readyData.Version
	// 基于 ready 事件，更新 session 信息
	c.Session.ID = readyData.SessionID
	c.Session.Shards.ShardID = readyData.Shard[0]
	c.Session.Shards.ShardCount = readyData.Shard[1]
	c.user = &WSUser{
		ID:       readyData.User.ID,
		Username: readyData.User.Username,
		Bot:      readyData.User.Bot,
	}
}

func ParseData(message []byte, target interface{}) error {
	data := gjson.Get(string(message), "d")
	return json.Unmarshal([]byte(data.String()), target)
}

func (c *Client) isHandleBuildIn(payload *PayloadCommon) bool {
	switch OPCode(payload.Op) {
	case WSHello: // 接收到 hello 后需要开始发心跳
		c.startHeartBeatTicker(payload.RawMessage)
	case WSHeartbeatAck: // 心跳 ack 不需要业务处理
	//case WSReconnect: // 达到连接时长，需要重新连接，此时可以通过 resume 续传原连接上的事件
	//	c.closeChan <- errs.ErrNeedReConnect
	//case WSInvalidSession: // 无效的 sessionLog，需要重新鉴权
	//	c.closeChan <- errs.ErrInvalidSession
	default:
		return false
	}
	return true
}

func (c *Client) startHeartBeatTicker(message []byte) {
	helloData := &HelloData{}
	if err := ParseData(message, helloData); err != nil {
		log.Printf("Error===================解析Hello数据失败")
	}
	// 根据 hello 的回包，重新设置心跳的定时器时间
	c.heartBeatTicker.Reset(time.Duration(helloData.HeartbeatInterval) * time.Millisecond)
}

func (c *Client) saveSeq(seq uint32) {
	if seq > 0 {
		c.Session.LastSeq = seq
	}
}
