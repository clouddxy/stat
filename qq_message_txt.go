package main

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"qiniupkg.com/x/log.v7"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// QQMessageTXT 导出 txt QQ消息
func QQMessageTXT() {

	session, err := mgo.Dial("10.30.23.45:7031")

	if err != nil {
		panic(err)
	}
	coll := session.DB("mars").C("qq_msg")
	defer session.Close()

	f, err := os.OpenFile("/home/qboxserver/tables/qq413-415.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Info(err)
	}
	defer f.Close()

	var messages []QQMessage
	start, _ := time.Parse("2006-01-02 15:04:05", "2017-04-13 00:00:00")
	end, _ := time.Parse("2006-01-02 15:04:05", "2017-04-19 00:00:00")
	err = coll.Find(bson.M{"createdat": bson.M{"$gte": start, "$lte": end}, "content": bson.M{"$regex": `^\-`}}).All(&messages)

	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, message := range messages {
		data, _ := json.Marshal(message)
		io.WriteString(f, string(data)+"\n")
	}

	err = coll.Find(bson.M{"createdat": bson.M{"$gte": start, "$lte": end}, "content": bson.M{"$regex": `^\+`}}).All(&messages)
	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, message := range messages {
		data, _ := json.Marshal(message)
		io.WriteString(f, string(data)+"\n")
	}
	log.Info("done")
}
