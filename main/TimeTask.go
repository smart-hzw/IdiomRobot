package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func DataUpdateworker() {
	ticker := time.NewTicker(20 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		dataToCache(5)
	}

}

func KeyExpireListeningworker(token string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		KeyExpireListening(token)
	}
}

// 监听redis中键过期
func KeyExpireListening(token string) {
	ttl, err := redisConn.TTL(context.Background(), IdiomSwitch).Result()
	if err != nil {

	}
	fmt.Println("===============定时任务", ttl)
	if ttl <= 10*time.Second && ttl >= 0 {
		//设置nextIdiom过期
		redisConn.Expire(context.Background(), Prelast, 1*time.Second)
		toCreate := &MessageToCreate{
			Content: string(TIMEOUT),
			//MsgID:   data.ID,
		}
		db := DB
		db.Exec("UPDATE idiom SET `status`=0")
		////发送消息
		url := "https://sandbox.api.sgroup.qq.com/channels/" + "655698385" + "/messages"
		jsonData, err := json.Marshal(toCreate)
		if err != nil {
			log.Printf("Info=========================json解析失败")
		}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		client := &http.Client{}

		req.Header.Add("Authorization", pre+token)
		req.Header.Add("X-Union-Appid", AppID)
		req.Header.Set("Content-Type", "application/json")
		// 发送请求并获取响应
		resp, err := client.Do(req)
		if err != nil {

		}
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println("==========================================键过期响应", string(body))
		var message *Message
		err2 := json.Unmarshal([]byte(body), &message)
		if err2 != nil {
			log.Fatalf("解析JSON出错: %v", err2)
		}
	}
}
