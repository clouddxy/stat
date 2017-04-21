package main

import (
	"encoding/csv"
	"os"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"qiniupkg.com/x/log.v7"
)

// QQMessage qq 消息
type QQMessage struct {
	CreatedAt   time.Time `bson:"createdat" json:"createdat"`                           // 创建时间
	Content     string    `bson:"content,omitempty" json:"content,omitempty"`           // 消息内容
	GroupNumber string    `bson:"group_number,omitempty" json:"group_number,omitempty"` // 群号
	GroupName   string    `bson:"groupname,omitempty" json:"groupname,omitempty"`       // 群名
	Day         string    `bson:"day,omitempty" json:"day,omitempty"`                   // 消息的日期
	NickName    string    `bson:"nick_name,omitempty" json:"nick_name,omitempty"`       // 昵称

}

// QQMessageCSV 导出 csv 格式 qqm essage
func QQMessageCSV() {

	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	coll := session.DB("mars").C("qq_msg")
	defer session.Close()

	f, err := os.OpenFile("/Users/dxy/Desktop/qqmsg/msg413-419.csv", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Info(err)
	}
	defer f.Close()

	var messagesPlus []QQMessage
	var messagesMinus []QQMessage
	start, _ := time.Parse("2006-01-02 15:04:05", "2017-04-13 00:00:00")
	end, _ := time.Parse("2006-01-02 15:04:05", "2017-04-20 00:00:00")
	err = coll.Find(bson.M{"createdat": bson.M{"$gte": start, "$lte": end}, "content": bson.M{"$regex": `^\-`}}).All(&messagesPlus)

	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Info(messagesPlus)
	err = coll.Find(bson.M{"createdat": bson.M{"$gte": start, "$lte": end}, "content": bson.M{"$regex": `^\+`}}).All(&messagesMinus)
	if err != nil {
		log.Error(err.Error())
		return
	}

	cwr := csv.NewWriter(f)
	cwr.Write([]string{"创建时间", "回复内容", "回复人", "群号"})

	statMessagesPlus := [][]string{}
	for _, statPlus := range messagesPlus {
		statMessagesPlus = append(statMessagesPlus, []string{
			statPlus.CreatedAt.Format("2006-01-02 15:04:05"),
			statPlus.Content,
			statPlus.NickName,
			statPlus.GroupNumber,
		})
	}
	statMessagesMinus := [][]string{}
	for _, statMinus := range messagesMinus {
		statMessagesMinus = append(statMessagesMinus, []string{
			statMinus.CreatedAt.Format("2006-01-02 15:04:05"),
			statMinus.Content,
			statMinus.NickName,
			statMinus.GroupNumber,
		})
	}
	cwr.WriteAll(statMessagesPlus)
	cwr.Write([]string{""})
	cwr.WriteAll(statMessagesMinus)
	cwr.Write([]string{""})

	for _, mPlus := range messagesPlus {
		cwr.Write([]string{mPlus.CreatedAt.Format("2006-01-02 15:04:05"), mPlus.Content, mPlus.Content, mPlus.NickName, mPlus.GroupNumber})
		for _, mMinus := range messagesMinus {
			if mPlus.GroupNumber == mMinus.GroupNumber {
				cwr.Write([]string{mPlus.CreatedAt.Format("2006-01-02 15:04:05"), mMinus.Content, mMinus.Content, mMinus.NickName, mMinus.GroupNumber})
			}
		}
		cwr.Write([]string{""})
	}

	log.Info("done")
}
