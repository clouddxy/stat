package main

import (
	"encoding/json"
	"io"
	"os"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"qiniupkg.com/x/log.v7"
)

// Messages qq 消息
type Messages struct {
	Groupname string `json:"groupname"`
	DayCounts []DayCount
}

// DayCount 每天的消息数
type DayCount struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

// QQDailyMessages 导出每天的 qq 消息数
func QQDailyMessages() {
	session, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		panic(err)
	}
	coll1 := session.DB("mars").C("qq_group")

	coll2 := session.DB("mars").C("qq_msg")

	f, err := os.OpenFile("/Users/dxy/Desktop/qqInfo.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Info(err)
	}

	groupNames := []string{}
	err = coll1.Find(bson.M{}).Distinct("groupname", &groupNames)
	log.Info(groupNames)

	for _, groupName := range groupNames {
		q1 := bson.M{"groupname": groupName}
		days := []string{}
		err = coll2.Find(q1).Distinct("day", &days)
		var messages Messages
		messages.Groupname = groupName
		for _, day := range days {
			n, _ := coll2.Find(bson.M{"groupname": groupName, "day": day}).Count()
			messages.DayCounts = append(messages.DayCounts, DayCount{
				Day:   day,
				Count: n,
			})
		}
		log.Info(messages)
		data, _ := json.Marshal(messages)
		io.WriteString(f, string(data)+"\n")
	}
}
