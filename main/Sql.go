package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mozillazg/go-pinyin"
	"log"
	"time"
)

type replyMeSSAGE string

const (
	NOTIDIOM    replyMeSSAGE = "这不是成语，请说出一个正确的成语参与我们的游戏哦"
	UNKNOWERROR replyMeSSAGE = "发生了未知错误，耐心等待，排查中~"
	HARDSEACH   replyMeSSAGE = "正在抓紧找成语的路上，请耐心等候~"
	PlAYGAME    replyMeSSAGE = "成语接龙游戏还未开始哦，你可以发送含有《成语接龙》的指令来开启游戏"
	ERRORPRE    replyMeSSAGE = "回答错误哦，没有衔接上一个成语.重新回答吧"
	TIMEOUT     replyMeSSAGE = "长时间没人回答，游戏结束了哦"
)

type IdiomStruct struct {
	ID    int    `db:"id"`
	First string `db:"first"`
	Last  string `db:"last"`
	Wrod  string `db:"wrod"`
}

var i = 0
var DB *sql.DB
var redisConn *redis.Client

func init() {
	sqlConn := "root:root@tcp(localhost:3306)/robot?charset=utf8&parseTime=True&loc=Local"
	DB, _ = sql.Open("mysql", sqlConn) // 使用本地时间，即东八区，北京时间
	// set pool params
	DB.SetMaxOpenConns(2000)
	DB.SetMaxIdleConns(1000)
	DB.SetConnMaxLifetime(time.Minute * 60) // mysql default conn timeout=8h, should < mysql_timeout
	err := DB.Ping()
	if err != nil {
		log.Fatalf("database init failed, err: ", err)
	}
	log.Println("mysql conn pool has initiated.")

	redisConn = redis.NewClient(&redis.Options{
		Addr:     "106.75.237.231:6379", // Redis地址
		Password: "",                    // 无密码
		DB:       0,                     // 使用默认DB
	})
}

func dataToCache(n int) {
	//连接数据集
	db := DB
	//var excutrSql = "WITH RankedOrders AS (SELECT id,word,`first`,last,`status` ROW_NUMBER() OVER(PARTITION BY last ORDER BY id) AS rn FROM idiom) SELECT id, word, `first` FROM (SELECT * from RankedOrders WHERE `status`!=1 ) as a WHERE rn <= 20;"
	var sqlString = "WITH RankedOrders AS (SELECT `id`, word, `first`,`status`, ROW_NUMBER() OVER(PARTITION BY `last` ORDER BY `id`) AS rn FROM idiom) SELECT `id`,`word`,`first` FROM (SELECT * from RankedOrders WHERE `status`!=1 )  as a WHERE rn <= ?;"
	//按结尾读音分类，每个类别取20个放入redis中
	//redis的数据结构是，key的格式是{"first"：{"id":"为所欲为","id":"为虎作伥"}}
	rows, err := db.Query(sqlString, n)
	if err != nil {
		log.Println("Error=======================窗口查询失败", err)
	}

	ctx := context.Background()
	pipe := redisConn.Pipeline()
	for rows.Next() {
		var id int
		var first string
		var word string
		if err := rows.Scan(&id, &word, &first); err != nil {
			log.Fatal(err)
		}

		set := pipe.SAdd(ctx, first, word)
		log.Println("Info=======================redis缓存数据写入中", i, word, set)
		i = i + 1
	}

	if _, err := pipe.Exec(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("Info=======================数据写入缓存成功")
}

func searchNextIdiom(preString string) (string, error) {
	//还要处理，多条消息带来的并发问题，要指定消息回复
	//1. 首先判断传入的字符是否是成语,统一的先去redis中查，再去mysql中查
	db := DB
	idiomStruct := selectIdiom(preString)
	if idiomStruct.Last == "" {
		return string(NOTIDIOM), nil
	}
	//此时该成语已经被用了，先写数据库，再删除缓存
	upDateStaus(preString)
	log.Println("Info=======================更新已传递的成语使用状态")
	result, err := redisConn.SRem(context.Background(), idiomStruct.First, preString).Result()
	if err != nil {
		log.Printf("Error===========================缓存删除失败： ", err)
	}
	log.Printf("Info===========================从缓存删除已遍历的成语： ", result)

	//获取到下一个成语的拼音后，开始去redis中查找，如果redis中有则，返回该结果，先去数据库更改该值状态，再从缓冲中删除
	pop := redisConn.SPop(context.Background(), idiomStruct.Last)
	s, err := pop.Result()
	if s != "" {

		s, err := pop.Result()
		if err != nil {
			log.Printf("Error===========================下一个成语获取失败： ", err)
		}
		idiomStruct.Wrod = s
		log.Printf("Info===========================redis获取到下一个成语应该为： ", idiomStruct.Wrod)
	} else {
		//去数据库查找

		raw1, err := db.Query("select `id`,`word` from idiom where `status`!=1 and `first`=?", idiomStruct.Last)
		if err != nil {
			log.Printf("Error===========================下一个成语获取失败： ", err)
			return string(HARDSEACH), nil
		}
		for raw1.Next() {
			err := raw1.Scan(&idiomStruct.ID, &idiomStruct.Wrod)
			if err != nil {
				log.Printf("Error===========================最后一个汉字拼音获取失败： ", err)
				return string(HARDSEACH), nil
			}
		}
		log.Printf("Info===========================mysql获取到下一个成语应该为： ", idiomStruct.Wrod)

		if err != nil {
			log.Fatal(err)
		}
		return idiomStruct.Wrod, nil
	}
	//标记该值已遍历
	upDateStaus(idiomStruct.Wrod)
	log.Println("Info=======================更新已传递的成语使用状态")
	return idiomStruct.Wrod, nil
}

func upDateStaus(word string) {
	db := DB
	// 执行更新
	res, err := db.Exec("UPDATE idiom SET `status` = 1 WHERE `word` = ?", word)
	if err != nil {
		log.Printf("Error===========================更新语句失效： ", err)
	}
	id, err := res.RowsAffected()
	log.Printf("Info===========================更新语句成功： ", id)
}

func selectIdiom(word string) IdiomStruct {
	db := DB
	raw, err := db.Query("select `id`,`first` ,`last`from idiom where `word`=?", word)
	if err != nil {
		log.Printf("Error===========================数据查询失败", err)
		return IdiomStruct{}
	}
	idiomStruct := IdiomStruct{}
	for raw.Next() {
		err = raw.Scan(&idiomStruct.ID, &idiomStruct.First, &idiomStruct.Last)
		if err != nil {
			log.Printf("Error===========================最后一个汉字拼音获取失败： ", err)
			return IdiomStruct{}
		}
	}
	log.Printf("Info===========================该成语的最后一个汉字拼音为： ", idiomStruct.Last)
	return idiomStruct
}

func linkComparePre(input string) bool {
	get := redisConn.Get(context.Background(), Prelast)
	log.Printf("+++++++++++++++++++++++++++++++", get)
	var last string
	err := get.Scan(&last)
	pylast := pinyin.LazyConvert(last, nil)
	if err != nil {
		log.Printf("Error===========================获取PreLast缓存失败： ", err)
		return true
	}
	//为空时是第一次，放行
	if get == nil {
		return true
	}
	pyfirst := pinyin.LazyConvert(input, nil)
	fmt.Println("=======", input, last)
	fmt.Println("=======", pyfirst[0], pylast[len(pylast)-1])
	if pyfirst[0] != pylast[len(pylast)-1] {
		return false
	}
	return true
}
