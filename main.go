package main

import (
	"IdiomRobot/TimeTask"
	"IdiomRobot/controller"
	"IdiomRobot/websocket"
	"log"
)

func main() {
	//初始化响应的数据
	controller.DataToCache(50)
	websocket.Setup()
	manager := websocket.New()
	//1.获取accessToken
	token, err := websocket.GetAccessToken()
	go TimeTask.DataUpdateworker()
	go TimeTask.KeyExpireListeningworker(token.AccessToken)
	if err != nil {
		log.Printf("Error=====================AccessToken获取失败 %v", err)
	}
	//获取wss链接
	wssInfo, err := websocket.SendMessageWithAuth(token)
	if err != nil {
		log.Printf("Error=====================访问GET /gateway/bot失败 %v", err)
	}

	//instent的产生
	intent := websocket.RegisterHandlers(
		// at 机器人事件，目前是在这个事件处理中有逻辑，会回消息，其他的回调处理都只把数据打印出来，不做任何处理
		controller.ATMessageEventHandlerImpl(token.AccessToken),

		controller.ReadyHandlerImpl(),
		// 互动事件
		//InteractionHandler(),
	)
	manager.SessionChanStart(wssInfo, token, &intent)
	//关闭连接
	defer controller.DB.Close()
	defer controller.RedisConn.Close()
}
