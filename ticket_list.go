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

// TicketInfo ticket info
type TicketInfo struct {
	UID           uint32    `json:"uid"`
	CreatedAt     time.Time `bson:"created_at"               json:"created_at,omitempty"`
	AssigneeStaff string    `bson:"assignee_staff,omitempty" json:"assignee_staff,omitempty"`
	Status        string    `json:"status"`
}

// TickeLists 导出 ticket 列表
func TickeLists() {
	session, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		panic(err)
	}
	coll := session.DB("mars").C("tickets")

	f, err := os.OpenFile("/Users/dxy/Desktop/ticketsInfo.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Info(err)
	}
	var tickets []TicketInfo
	err = coll.Find(bson.M{}).All(&tickets)
	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, ticket := range tickets {
		log.Info(ticket)
		if ticket.UID != 0 {
			data, _ := json.Marshal(ticket)
			io.WriteString(f, string(data)+"\n")
		}
	}
	log.Info("Done")
}
