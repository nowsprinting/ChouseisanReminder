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
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

// 購読者エンティティ
type subscriber struct {
	DisplayName string
	MID         string
}

// 購読者の追加・削除ログを保存するエンティティ
type logSubscriber struct {
	DisplayName string
	MID         string
	OpType      string
	AddTime     time.Time
}

func init() {
	http.HandleFunc("/line/callback", lineCallback)
	http.HandleFunc("/task/addfriend", addFriend)
	http.HandleFunc("/task/removefriend", removeFriend)
	http.HandleFunc("/task/analyzecommand", analyzeCommand)
	http.HandleFunc("/cron/crawlchouseisan", crawlChouseisan)
	http.HandleFunc("/", usage)
}

func createBotClient(c context.Context) (bot *linebot.Client, err error) {
	var (
		channelSecret = os.Getenv("LINE_CHANNEL_SECRET")
		channelToken  = os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
	)

	bot, err = linebot.New(channelSecret, channelToken, linebot.WithHTTPClient(urlfetch.Client(c))) //Appengineのurlfetchを使用する
	if err != nil {
		log.Errorf(c, "Error occurred at create linebot client: %v", err)
		return bot, err
	}
	return bot, nil
}

/**
 * LINEからのコールバックをハンドリング
 */
func lineCallback(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	bot, err := createBotClient(c)
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
		source := event.Source
		if event.Type == linebot.EventTypeFollow {
			task := taskqueue.NewPOSTTask("/task/addfriend", url.Values{
				"mid": {source.UserID},
			})
			taskqueue.Add(c, task, "default")

		} else if event.Type == linebot.EventTypeUnfollow {
			task := taskqueue.NewPOSTTask("/task/removefriend", url.Values{
				"mid": {source.UserID},
			})
			taskqueue.Add(c, task, "default")

		} else if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if source.Type == linebot.EventSourceTypeUser {
					//テキストメッセージ受信 from User
					task := taskqueue.NewPOSTTask("/task/analyzecommand", url.Values{
						"token": {event.ReplyToken},
						"mid":   {source.UserID},
						"text":  {message.Text},
					})
					taskqueue.Add(c, task, "default")

				} else if source.Type == linebot.EventSourceTypeRoom {
					//テキストメッセージ受信 from Room
					task := taskqueue.NewPOSTTask("/task/analyzecommand", url.Values{
						"token": {event.ReplyToken},
						"mid":   {source.RoomID},
						"text":  {message.Text},
					})
					taskqueue.Add(c, task, "default")

				} else if source.Type == linebot.EventSourceTypeGroup {
					//テキストメッセージ受信 from Group
					task := taskqueue.NewPOSTTask("/task/analyzecommand", url.Values{
						"token": {event.ReplyToken},
						"mid":   {source.GroupID},
						"text":  {message.Text},
					})
					taskqueue.Add(c, task, "default")
				}
			}

		} else {
			//未サポートのイベントタイプ
			task := taskqueue.NewPOSTTask("/task/analyzecommand", url.Values{
				"token": {event.ReplyToken},
				"mid":   {"mid"},
				"text":  {"received event: " + string(event.Type)},
			})
			taskqueue.Add(c, task, "default")

		}
	}
}

/**
 * 送信者の情報を取得する
 */
func getSenderName(c context.Context, bot *linebot.Client, from string) string {
	senderProfile, err := bot.GetProfile(from).Do()
	if err != nil {
		log.Errorf(c, "Error occurred at get sender profile. from: %v, err: %v", from, err)
		return from
	}
	return senderProfile.DisplayName
}

/**
 * 購読者全員にメッセージを送信
 */
func sendToAll(c context.Context, bot *linebot.Client, message string) error {

	//データストアから購読者のMIDを取得
	q := datastore.NewQuery("Subscriber")
	var subscribers []subscriber
	if _, err := q.GetAll(c, &subscribers); err != nil {
		log.Errorf(c, "Error occurred at get-all from datastore. err: %v", err)
		return err
	}

	//全員に送信
	for _, current := range subscribers {
		if _, err := bot.PushMessage(current.MID, linebot.NewTextMessage(message)).Do(); err != nil {
			log.Errorf(c, "Error occurred at send message: %v", err)
			return err
		}
	}

	return nil
}

/**
 * 友だち追加（データストアに購読者として登録）
 */
func addFriend(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	bot, err := createBotClient(c)
	if err != nil {
		return
	}

	//購読者プロファイルを取得
	mid := r.FormValue("mid")
	senderName := getSenderName(c, bot, mid)

	//購読者を保存
	entity := subscriber{
		DisplayName: senderName,
		MID:         mid,
	}
	key := datastore.NewKey(c, "Subscriber", mid, 0, nil)
	if _, err := datastore.Put(c, key, &entity); err != nil {
		log.Errorf(c, "Error occurred at put subcriber to datastore. mid:%v, err: %v", mid, err)
		return
	}

	//ログエントリを追加
	logEntity := logSubscriber{
		DisplayName: senderName,
		MID:         mid,
		OpType:      string(linebot.EventTypeFollow),
		AddTime:     time.Now(),
	}
	logKey := datastore.NewIncompleteKey(c, "LogSubscriber", nil)
	if _, err := datastore.Put(c, logKey, &logEntity); err != nil {
		log.Errorf(c, "Error occurred at put log-subcriber to datastore. mid:%v, err: %v", mid, err)
		return
	}

	sendToAll(c, bot, senderName+"さんが購読を開始しました。いらっしゃいませ！！")
}

/**
 * 友だち削除（データストアから削除）
 */
func removeFriend(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	bot, err := createBotClient(c)
	if err != nil {
		return
	}

	//購読者プロファイルを取得
	mid := r.FormValue("mid")
	senderName := getSenderName(c, bot, mid)

	//購読者を削除
	key := datastore.NewKey(c, "Subscriber", mid, 0, nil)
	if err := datastore.Delete(c, key); err != nil {
		log.Errorf(c, "Error occurred at delete subcriber to datastore. mid:%v, err: %v", mid, err)
		return
	}

	//ログエントリを追加
	logEntity := logSubscriber{
		DisplayName: senderName,
		MID:         mid,
		OpType:      string(linebot.EventTypeUnfollow),
		AddTime:     time.Now(),
	}
	logKey := datastore.NewIncompleteKey(c, "LogSubscriber", nil)
	if _, err := datastore.Put(c, logKey, &logEntity); err != nil {
		log.Errorf(c, "Error occurred at put log-subcriber to datastore. mid:%v, err: %v", mid, err)
		return
	}

	sendToAll(c, bot, senderName+"さんが購読を解除しました。さようなら！")
}

/**
 * チャットコマンド解析（コマンドに応じたメッセージを送信）
 */
func analyzeCommand(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	bot, err := createBotClient(c)
	if err != nil {
		return
	}
	mid := r.FormValue("mid")
	text := r.FormValue("text")

	//TODO: コマンドに応じて応答を変える

	//default: 全員にブロードキャスト
	senderName := getSenderName(c, bot, mid)
	sendToAll(c, bot, senderName+"さんより\n「"+text+"」")
}

/**
 * Usageを表示
 */
func usage(w http.ResponseWriter, r *http.Request) {
	response := template.Must(template.ParseFiles("templates/usage.html"))
	response.Execute(w, struct {
		Hash string //調整さんのイベントハッシュ
	}{
		Hash: os.Getenv("CHOUSEISAN_EVENT_HASH"),
	})
}
