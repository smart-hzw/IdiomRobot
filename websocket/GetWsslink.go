package websocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	AppID        = "102141917"
	ClientSecret = "5Ql6Rm7SoAWsEawIf2Pm9WtHf3RpDbzO" //正确应该走数据库，加密读取
)

// 请求token的结构体
type Tokenrequest struct {
	AppId        string `json:"appId"`
	ClientSecret string `json:"clientSecret"`
}

func GetAccessToken() (*Token, error) {
	// 构建鉴权请求的URL和参数（这里简化为POST表单数据，实际可能使用JSON或其他格式）
	// 注意：这里只是示例，实际URL和参数应根据API文档确定
	authURL := "https://bots.qq.com/app/getAppAccessToken"
	client := &http.Client{}

	tokenrequest := Tokenrequest{
		AppId:        AppID,
		ClientSecret: ClientSecret,
	}

	jsonData, err := json.Marshal(tokenrequest)

	req, err := http.NewRequest("POST", authURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
	}

	// 设置请求头，指明我们发送的是JSON格式的数据
	req.Header.Set("Content-Type", "application/json")

	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	// 打印响应状态码和响应体
	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Println("Response Body:", string(body))
	var token Token
	err2 := json.Unmarshal([]byte(body), &token)
	if err2 != nil {
		log.Fatalf("解析JSON出错: %v", err2)
	}

	return &token, nil
}

func SendMessageWithAuth(token *Token) (*WssInfo, error) {
	// 将请求体转换为JSON格式
	url := "https://sandbox.api.sgroup.qq.com/gateway/bot"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// 处理错误
		fmt.Println("Error creating request:", err)
	}

	// 设置请求头
	req.Header.Add("Authorization", "QQBot "+token.AccessToken)
	req.Header.Add("X-Union-Appid", AppID)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析响应体
	var wssInfo WssInfo
	err1 := json.Unmarshal([]byte(respBody), &wssInfo)
	if err != nil {
		return nil, err1
	}
	fmt.Println("Response Body:", wssInfo.URL)

	return &wssInfo, nil
}
