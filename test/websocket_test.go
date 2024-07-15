package test

import (
	"IdiomRobot/dto"
	"IdiomRobot/websocket"
	"log"
	"testing"
)

//var token *tmp.Token
//var wssInfo *tmp.WssInfo

func Test_webSocket(t *testing.T) {
	token, err := websocket.GetAccessToken()
	wssInfo, err := websocket.SendMessageWithAuth(token)
	log.Printf("=======================%v%v", token.AccessToken, wssInfo.URL)
	if err != nil {
		log.Print("=================", err)
	}
	t.Run(
		"at message", func(t *testing.T) {
			var message websocket.ATMessageEventHandler = func(event *dto.PayloadCommon, data *dto.ATMessageData) error {
				log.Println(event, data)
				return nil
			}
			intent := websocket.RegisterHandlers(message)

			manager := websocket.New()

			manager.SessionChanStart(wssInfo, token, &intent)
		},
	)
}
