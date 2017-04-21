package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"qiniupkg.com/x/log.v7"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Satisfaction info
type Satisfaction struct {
	ID                 int                `bson:"id" json:"id"`                                             // 创建工单时系统自动分配
	SatisfactionRating SatisfactionRating `bson:"satisfaction_rating" json:"satisfaction_rating"`           // 满意度评价
	AssigneeStaff      string             `bson:"assignee_staff,omitempty" json:"assignee_staff,omitempty"` // 七牛员工
}

// SatisfactionRating ticket satisfaction rating
type SatisfactionRating struct {
	ID               int    `bson:"id" json:"id"`
	Score            string `bson:"score" json:"score"`
	CurrentService   int    `bson:"current_service" json:"current_service"`
	SystemUsability  int    `bson:"system_usability" json:"system_usability"`
	ProductSatisfact int    `bson:"product_satisfact" json:"product_satisfact"`
	ServiceQuality   int    `bson:"service_quality" json:"service_quality"`
	ResponseSpeed    int    `bson:"response_speed" json:"response_speed"`
	ServiceAttitude  int    `bson:"service_attitude" json:"service_attitude"`
	Comment          string `bson:"comment" json:"comment"`
}

// SatisfactionCSV 导出 CSV 格式的 ticket 数据
func SatisfactionCSV() {

	session, err := mgo.Dial("10.30.23.45:7031")

	if err != nil {
		panic(err)
	}
	coll := session.DB("mars").C("tickets")
	f, err := os.OpenFile("/home/qboxserver/tables/tickets_satisfaction_rating.csv", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Info(err)
	}

	cwr := csv.NewWriter(f)
	cwr.Write([]string{"工单号", "本次服务总体体验", "工单系统是否方便使用", "产品满意度", "工单服务质量", "工单回复速度", "工单服务态度", "评论", "受理客服"})
	var satisfactions []Satisfaction

	query := bson.M{"satisfaction_rating.comment": bson.M{"$ne": ""}, "satisfaction_rating.score": ""}
	err = coll.Find(query).All(&satisfactions)

	if err != nil {
		log.Error(err.Error())
		return
	}
	for _, satisfaction := range satisfactions {
		cwr.Write([]string{
			fmt.Sprintf("%d", satisfaction.ID),
			fmt.Sprintf("%d", satisfaction.SatisfactionRating.CurrentService),
			fmt.Sprintf("%d", satisfaction.SatisfactionRating.SystemUsability),
			fmt.Sprintf("%d", satisfaction.SatisfactionRating.ProductSatisfact),
			fmt.Sprintf("%d", satisfaction.SatisfactionRating.ServiceQuality),
			fmt.Sprintf("%d", satisfaction.SatisfactionRating.ResponseSpeed),
			fmt.Sprintf("%d", satisfaction.SatisfactionRating.ServiceAttitude),
			satisfaction.SatisfactionRating.Comment,
			satisfaction.AssigneeStaff,
		})
	}
	log.Info("Done")
}
