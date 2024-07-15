package controller

import (
	"IdiomRobot/dto"
	"IdiomRobot/websocket"
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

// ReadyHandler 自定义 ReadyHandler 感知连接成功事件
func ReadyHandlerImpl() websocket.ReadyHandler {
	return func(event *dto.PayloadCommon, data *dto.ReadyData) {
		log.Println("ready event receive: ", data)
	}
}

// ATMessageEventHandler 实现处理 at 消息的回调
func ATMessageEventHandlerImpl(accesstoken string) websocket.ATMessageEventHandler {
	return func(event *dto.PayloadCommon, data *dto.ATMessageData) error {
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

func ProcessMessage(input string, data *dto.ATMessageData, accesstoken string) error {
	var replyContent string
	var expire_time = 2 * time.Minute
	//回复消息
	//timestamp := data.Timestamp
	log.Println("Info===========================前台收到： ", input, len(input))
	re, err := regexp.Compile(`.*成语接龙.*`)
	get := RedisConn.Get(context.Background(), IdiomSwitch)
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
			set := RedisConn.Set(context.Background(), IdiomSwitch, true, expire_time)
			log.Println("Info===========================成语接龙标识写入缓存中： ", set)
		}
	} else {
		idiom := SelectIdiom(input)
		if flag == true { //进入成语接龙判断
			RedisConn.Set(context.Background(), IdiomSwitch, true, expire_time)
			RedisConn.Expire(context.Background(), Prelast, expire_time)
			if idiom.Last == "" {
				replyContent = string(NOTIDIOM)
			} else {
				comparePre := LinkComparePre(input)
				log.Printf("+++++++++++++++++++++++++++++++", comparePre)
				if comparePre {
					nextIdiom, err := SearchNextIdiom(input)
					if err != nil {
						log.Printf("Error=====================思考中，遇到点挫折 %v", err)
					}
					replyContent = nextIdiom
					set := RedisConn.Set(context.Background(), Prelast, nextIdiom, expire_time)
					log.Println("Info===========================记录上一个成语末尾汉字： ", set)
				} else {
					replyContent = string(ERRORPRE)
				}
			}
		} else {
			if idiom.Last != "" {
				replyContent = string(PlAYGAME)
			} else {
				replyContent = "Hello World" + dto.Emoji(307)
			}
		}
	}
	log.Printf("Info===========================回应消息为： ", replyContent)
	toCreate := &dto.MessageToCreate{
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

func replyMessage(channelID string, msg *dto.MessageToCreate, accesstoken string) (*dto.Message, error) {
	fmt.Println("=======================================================消息分割线", channelID)
	log.Printf("Info=====================机器人回复消息中……")
	url := "https://sandbox.api.sgroup.qq.com/channels/" + channelID + "/messages"
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Info=========================json解析失败")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	client := &http.Client{}

	req.Header.Add("Authorization", websocket.Pre+accesstoken)
	req.Header.Add("X-Union-Appid", websocket.AppID)
	req.Header.Set("Content-Type", "application/json")
	// 发送请求并获取响应
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	var message *dto.Message
	log.Printf("Info=====================机器人回复完成", resp.Status)
	err2 := json.Unmarshal([]byte(body), &message)
	if err2 != nil {
		log.Fatalf("解析JSON出错: %v", err2)
	}
	return message, nil
}
