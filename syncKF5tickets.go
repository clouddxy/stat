package main

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"labix.org/v2/mgo"

	"qbox.us/biz/component/client"
	"qbox.us/mgo2"
	"qiniupkg.com/x/log.v7"

	"github.com/qiniu/xlog.v1"

	"mars.qiniu.com/collections"
	"mars.qiniu.com/config"
	"mars.qiniu.com/models"
	"mars.qiniu.com/ticket"
	"mars.qiniu.com/util/httpauth"
	"mars.qiniu.com/util/rpc"
	"mars.qiniu.com/util/timeutil"
)

// TicketClient ticket client
var TicketClient *Client

// DevClient developer client
var DevClient *DeveloperClient
var coll *mgo.Collection

// Client ticket client
type Client struct {
	Host string
	Conn rpc.Client
}

func newTicketClient(host, user, passwd string) *Client {
	tr := httpauth.NewBasicTransport(user, passwd, nil)
	return &Client{
		Host: host,
		Conn: rpc.Client{Client: &http.Client{Transport: tr}},
	}
}

// DeveloperClient struct
type DeveloperClient struct {
	Host string
	Conn rpc.Client
}

// newDeveloperClient 返回开发者信息的 Client
func newDeveloperClient(host, user, passwd string) (devClient *DeveloperClient, err error) {

	adminOauth := client.NewAdminOAuth(host, http.DefaultTransport)
	token, code, err := adminOauth.ExchangeByPassword(user, passwd)
	if code != http.StatusOK || err != nil {
		log.Error(err.Error())
		return nil, err
	}

	adminOauth.Token = token
	return &DeveloperClient{
		Host: host,
		Conn: rpc.Client{Client: &http.Client{Transport: adminOauth}},
	}, nil
}

// KF5Ticket ticket struct
type KF5Ticket struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	RequesterID int    `json:"requester_id"`
	CreatedAt   CTime  `json:"created_at"`
}

// User ticket user struct
type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// CTime ctime struct
type CTime struct {
	time.Time
}

// DeveloperReply struct
type DeveloperReply struct {
	Code int               `json:"code"`
	Data *models.Developer `json:"data"`
}

func init() {
	TicketClient = newTicketClient("https://qiniu.kf5.com", "lvguihua@qiniu.com/token", "c680240e889e801c17d376a34ede2a")
	DevClient, _ = newDeveloperClient("https://acc.qbox.me", "qcos-qboxacc@qiniu.com", "BgVfeD3uqp2QS78Tnqku4r67P3c6849MHE5CHNCJ")
	if err := config.Init(); err != nil {
		log.Fatal("Load config file error:", err.Error())
	}
	session := mgo2.Open(&mgo2.Config{
		Host: config.Mongo.Host,
		DB:   config.Mongo.Name,
		Mode: config.Mongo.Mode,
		Coll: "tickets",
	})
	coll = session.Coll

	// Init collection mgrs
	if err := collections.Init(config.Mongo); err != nil {
		log.Fatal("Mgrs init fatal:", err.Error())
	}

	// 初始化工单模块
	if err := ticket.Init(config.Mongo); err != nil {
		log.Fatal("Ticket module init fatal:", err.Error())
	}

}

// DeveloperByEmail get developer by email
func (c *DeveloperClient) DeveloperByEmail(log *xlog.Logger, email string) (developer *models.Developer, err error) {
	url := "https://api.qiniu.com/api/developer?email=" + email
	rel := &DeveloperReply{}
	err = c.Conn.GetCall(log, rel, url)
	if err != nil {
		return nil, err
	}
	developer = rel.Data
	if developer == nil {
		return nil, errors.New("Not Found")
	}
	return
}

// UnmarshalJSON UnmarshalJSON
func (ct *CTime) UnmarshalJSON(b []byte) (err error) {
	bs := string(b)
	if bs == "null" {
		bs = "1970-01-01 00:00:00"
	}
	reg := regexp.MustCompile(`^"(.*)"$`)
	bs = reg.ReplaceAllString(bs, "${1}")
	ctLayout := "2006-01-02 15:04:05"
	ct.Time, err = time.Parse(ctLayout, bs)
	return
}

func main() {
	log := xlog.NewDummy()
	ticketClient := newTicketClient("https://qiniu.kf5.com", "xxxx", "xxxx")

	kf5Ticket := &struct {
		Ticket KF5Ticket `json:"ticket"`
	}{}

	user := &struct {
		User User `json:"user"`
	}{}

	ticketIDs := []int{100000}
	log.Info(ticketIDs)

	ch := make(chan int, 1)
	wg := sync.WaitGroup{}
	for _, ticketID := range ticketIDs {

		ch <- ticketID
		wg.Add(1)
		go func(tickekID int) {
			log.Info(tickekID)
			if err := ticketClient.Conn.GetCall(log, &kf5Ticket, fmt.Sprintf("https://qiniu.kf5.com/apiv2/tickets/%d.json", tickekID)); err != nil {
				log.Error(err.Error())
				return
			}
			if err := ticketClient.Conn.GetCall(log, &user, fmt.Sprintf("https://qiniu.kf5.com/apiv2/users/%d.json", kf5Ticket.Ticket.RequesterID)); err != nil {
				log.Error(err.Error())
				return
			}

			developer, err := DevClient.DeveloperByEmail(log, user.User.Email)
			if err != nil {
				log.Error(err.Error())
				wg.Done()
				<-ch
				return
			}

			if developer == nil {
				wg.Done()
				<-ch
				return
			}

			t := &ticket.Ticket{}
			t.UID = developer.UID
			t.Status = "open"
			t.Title = kf5Ticket.Ticket.Title
			t.Description = kf5Ticket.Ticket.Description
			t.CreatedAt = kf5Ticket.Ticket.CreatedAt.Time
			t.UpdatedAt = kf5Ticket.Ticket.CreatedAt.Time
			t.Product = "Portal"
			t.Category = "财务和计费｜账单和费用问题"
			t.IsPrivate = "public"
			// 获取自增 ID
			if t.ID, err = collections.CounterMgr.GetID("Ticket"); err != nil {
				return
			}
			if err = coll.Insert(t); err != nil { // 创建工单
				return
			}

			comment := &ticket.Comment{
				TicketID:    t.ID,
				Content:     t.Description,
				HTMLContent: t.Description,
				CreatedAt:   timeutil.ShortTime(t.CreatedAt),
				Public:      true,
				UID:         t.UID,
				Email:       t.Email,
				AuthorName:  user.User.Name,
			}
			if err = ticket.CommentMgr.Create(comment); err != nil {
				return
			}
			wg.Done()
			<-ch
		}(ticketID)
	}

	close(ch)
	wg.Wait()
}
