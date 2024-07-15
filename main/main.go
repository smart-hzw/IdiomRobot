package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// 成语接龙游戏开启标识 缓存key
var IdiomSwitch = "IdiomSwitch"

// 记录上一个成语last 缓存
var Prelast = "PreLast"

func main() {
	//初始化响应的数据
	dataToCache(50)
	Setup()
	manager := New()
	//1.获取wss链接
	token, err := getAccessToken()
	go DataUpdateworker()
	go KeyExpireListeningworker(token.AccessToken)
	if err != nil {
		log.Printf("Error=====================AccessToken获取失败 %v", err)
	}
	wssInfo, err := sendMessageWithAuth(token)
	if err != nil {
		log.Printf("Error=====================访问GET /gateway/bot失败 %v", err)
	}

	//instent的产生
	intent := RegisterHandlers(
		// at 机器人事件，目前是在这个事件处理中有逻辑，会回消息，其他的回调处理都只把数据打印出来，不做任何处理
		ATMessageEventHandlerImpl(token.AccessToken),

		ReadyHandlerImpl(),
		// 互动事件
		//InteractionHandler(),
	)
	manager.SessionChanStart(wssInfo, token, &intent)

	//关闭连接
	defer DB.Close()
	defer redisConn.Close()

	//searchNextIdiom("不三不四")

}

// ReadyHandler 自定义 ReadyHandler 感知连接成功事件
func ReadyHandlerImpl() ReadyHandler {
	return func(event *PayloadCommon, data *ReadyData) {
		log.Println("ready event receive: ", data)
	}
}

// ATMessageEventHandler 实现处理 at 消息的回调
func ATMessageEventHandlerImpl(accesstoken string) ATMessageEventHandler {
	return func(event *PayloadCommon, data *ATMessageData) error {
		input := strings.ToLower(ETLInput(data.Content))
		return ProcessMessage(input, data, accesstoken)
	}
}

var atRE = regexp.MustCompile(`<@!\d+>`)

const spaceCharSet = " \u00A0"

func ETLInput(input string) string {
	etlData := string(atRE.ReplaceAll([]byte(input), []byte("")))
	etlData = strings.Trim(etlData, spaceCharSet)
	return etlData
}

func ProcessMessage(input string, data *ATMessageData, accesstoken string) error {
	var replyContent string
	var expire_time = 2 * time.Minute
	//回复消息
	//timestamp := data.Timestamp
	log.Println("Info===========================前台收到： ", input, len(input))
	re, err := regexp.Compile(`.*成语接龙.*`)
	get := redisConn.Get(context.Background(), IdiomSwitch)
	var flag bool
	err = get.Scan(&flag)
	fmt.Println("=====================================flag", flag)
	if err != nil {
		log.Printf("Error=====================成语接龙标识缓存获取失败 %v", err)
	}
	if re.MatchString(input) {
		if flag == true {
			replyContent = "已经在玩成语接龙了呢，要来一起参与吗"
		} else {
			//触发接龙，将触发标记放入缓存中
			replyContent = "欢迎参加成语接龙游戏，请你开始说出你的成语"
			set := redisConn.Set(context.Background(), IdiomSwitch, true, expire_time)
			log.Println("Info===========================成语接龙标识写入缓存中： ", set)
		}
	} else {
		idiom := selectIdiom(input)
		if flag == true { //进入成语接龙判断
			redisConn.Set(context.Background(), IdiomSwitch, true, expire_time)
			redisConn.Expire(context.Background(), Prelast, expire_time)
			if idiom.Last == "" {
				replyContent = string(NOTIDIOM)
			} else {
				comparePre := linkComparePre(input)
				log.Printf("+++++++++++++++++++++++++++++++", comparePre)
				if comparePre {
					nextIdiom, err := searchNextIdiom(input)
					if err != nil {
						log.Printf("Error=====================思考中，遇到点挫折 %v", err)
					}
					replyContent = nextIdiom
					set := redisConn.Set(context.Background(), Prelast, nextIdiom, expire_time)
					log.Println("Info===========================记录上一个成语末尾汉字： ", set)
				} else {
					replyContent = string(ERRORPRE)
				}
			}
		} else {
			if idiom.Last != "" {
				replyContent = string(PlAYGAME)
			} else {
				replyContent = "Hello World" + Emoji(307)
			}
		}
	}
	log.Printf("Info===========================回应消息为： ", replyContent)
	toCreate := &MessageToCreate{
		Content: replyContent,
		MsgID:   data.ID,
	}
	message, err := replyMessage(data.ChannelID, toCreate, accesstoken)
	if err != nil {
		log.Printf("Error=====================AT消息回复失败 %v", err)
	}
	log.Printf("Info=====================AT消息回复成功 %v", message)
	return nil
}

func replyMessage(channelID string, msg *MessageToCreate, accesstoken string) (*Message, error) {
	fmt.Println("=======================================================消息分割线", channelID)
	log.Printf("Info=====================机器人回复消息中……")
	url := "https://sandbox.api.sgroup.qq.com/channels/" + channelID + "/messages"
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Info=========================json解析失败")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	client := &http.Client{}

	req.Header.Add("Authorization", pre+accesstoken)
	req.Header.Add("X-Union-Appid", AppID)
	req.Header.Set("Content-Type", "application/json")
	// 发送请求并获取响应
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	var message *Message
	log.Printf("Info=====================机器人回复完成", resp.Status)
	err2 := json.Unmarshal([]byte(body), &message)
	if err2 != nil {
		log.Fatalf("解析JSON出错: %v", err2)
	}
	return message, nil
}
