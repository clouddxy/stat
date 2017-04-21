package main

import (
	"encoding/json"
	"io"
	"os"

	"qiniupkg.com/x/log.v7"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// SatisfactionTXT 导出 TXT 格式的满意度
func SatisfactionTXT() {

	session, err := mgo.Dial("127.0.0.1:27017")

	if err != nil {
		panic(err)
	}
	coll := session.DB("mars").C("tickets")
	f, err := os.OpenFile("/Users/dxy/Desktop/tickets/tickets_satisfaction.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)

	if err != nil {
		log.Info(err)
	}

	var satisfactions []Satisfaction
	query := bson.M{"satisfaction_rating.comment": bson.M{"$ne": ""}, "satisfaction_rating.score": ""}
	err = coll.Find(query).All(&satisfactions)

	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, satisfaction := range satisfactions {

		log.Info(satisfaction)
		data, _ := json.Marshal(satisfaction)
		io.WriteString(f, string(data)+"\n")

	}
	log.Info("Done")
}
