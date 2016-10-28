package main

import (
	"html/template"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/line/line-bot-sdk-go/linebot"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

// 購読者エンティティ（keyはMID）
type subscriber struct {
	DisplayName    string // 表示名（取得できるのはユーザの場合のみ）
	MID            string // ユーザ/グループ/ルームのid
	ChouseisanHash string // リマインド対象の調整さんのハッシュ
	RemindBefore   int    // イベントの何日前にリマインド処理を行なうか。デフォルトは3日
	RemindTime     int    // 何時にリマインド処理を行なうか（日本時間）。デフォルトは8:00
}

// 購読者の追加・削除ログを保存するエンティティ
type logSubscriber struct {
	DisplayName string
	MID         string
	EventType   string
	AddTime     time.Time
}

func init() {
	http.HandleFunc("/line/callback", lineCallback)
	http.HandleFunc("/task/join", join)
	http.HandleFunc("/task/leave", leave)
	http.HandleFunc("/task/commandanalyze", commandAnalyze)
	http.HandleFunc("/cron/crawlchouseisan", crawlChouseisan)
	http.HandleFunc("/", usage)
}

func createBotClient(c context.Context, client *http.Client) (bot *linebot.Client, err error) {
	var (
		channelSecret = os.Getenv("LINE_CHANNEL_SECRET")
		channelToken  = os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	)

	bot, err = linebot.New(channelSecret, channelToken, linebot.WithHTTPClient(client)) //Appengineのurlfetchを使用する
	if err != nil {
		log.Errorf(c, "Error occurred at create linebot client: %v", err)
		return bot, err
	}
	return bot, nil
}

/**
 * Get event sender's id
 */
func getSenderID(c context.Context, event *linebot.Event) string {
	switch event.Source.Type {
	case linebot.EventSourceTypeGroup:
		return event.Source.GroupID
	case linebot.EventSourceTypeRoom:
		return event.Source.RoomID
	case linebot.EventSourceTypeUser:
		return event.Source.UserID
	}
	log.Warningf(c, "Can not get sender id. type: %v", event.Source.Type)
	return ""
}

/**
 * LINEからのコールバックをハンドリング
 */
func lineCallback(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	bot, err := createBotClient(c, urlfetch.Client(c))
	if err != nil {
		return
	}

	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			log.Warningf(c, "Linebot request status: 400")
			w.WriteHeader(400)
		} else {
			log.Warningf(c, "linebot request status: 500\n\terror: %v", err)
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeFollow, linebot.EventTypeJoin:
			task := taskqueue.NewPOSTTask("/task/join", url.Values{
				"mid":        {getSenderID(c, event)},
				"type":       {string(event.Type)},
				"replyToken": {event.ReplyToken},
			})
			taskqueue.Add(c, task, "default")

		case linebot.EventTypeUnfollow, linebot.EventTypeLeave:
			task := taskqueue.NewPOSTTask("/task/leave", url.Values{
				"mid":  {getSenderID(c, event)},
				"type": {string(event.Type)},
			})
			taskqueue.Add(c, task, "default")

		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if message.Text[0:1] == "/" {
					task := taskqueue.NewPOSTTask("/task/commandanalyze", url.Values{
						"mid":        {getSenderID(c, event)},
						"replyToken": {event.ReplyToken},
						"text":       {message.Text[1:]},
					})
					taskqueue.Add(c, task, "default")
				}
			}

		default:
			log.Debugf(c, "Unsupported event type. type: %v", event.Type)
		}
	}
}

/**
 * Usageを表示
 */
func usage(w http.ResponseWriter, r *http.Request) {
	response := template.Must(template.ParseFiles("templates/usage.html"))
	response.Execute(w, nil)
}
