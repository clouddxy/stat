package main

import (
	"encoding/json"
	"io"
	"os"

	"qiniupkg.com/x/log.v7"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// QQGroupInfo ticket info
type QQGroupInfo struct {
	UID       string `bson:"uid,omitempty"             json:"uid,omitempty"`       // 群用户的UID
	GroupName string `bson:"groupname,omitempty"       json:"groupname,omitempty"` // 群名
	QQ        string `bson:"qq,omitempty"              json:"qq,omitempty"`        // 群号码
}

// QQGroupList 导出 qq 列表
func QQGroupList() {
	session, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		panic(err)
	}
	coll := session.DB("mars").C("qq_group")

	f, err := os.OpenFile("/Users/dxy/Desktop/qq_groupInfo.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Info(err)
	}
	var qqGroups []QQGroupInfo
	err = coll.Find(bson.M{}).All(&qqGroups)
	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, qqGroup := range qqGroups {
		log.Info(qqGroup)
		data, _ := json.Marshal(qqGroup)
		io.WriteString(f, string(data)+"\n")
	}
	log.Info("Done")
}
