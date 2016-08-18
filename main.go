package main

import (
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/line/line-bot-sdk-go/linebot"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
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
	OpType      linebot.OpType
	AddTime     time.Time
}

func init() {
	http.HandleFunc("/line/callback", lineCallback)
	http.HandleFunc("/", usage)
}

/**
 * LINEからのコールバックをハンドリング
 */
func lineCallback(w http.ResponseWriter, r *http.Request) {
	var (
		channelID     int64
		channelSecret = os.Getenv("LINE_CHANNEL_SECRET")
		channelMID    = os.Getenv("LINE_CHANNEL_MID")
		err           error
	)

	// App engine Content
	c := appengine.NewContext(r)

	// Setup bot client
	channelID, err = strconv.ParseInt(os.Getenv("LINE_CHANNEL_ID"), 10, 64)
	if err != nil {
		log.Errorf(c, "Error occurred at parse `LINE_CHANNEL_ID`: %v", err)
		return
	}
	bot, err := linebot.NewClient(channelID, channelSecret, channelMID, linebot.WithHTTPClient(urlfetch.Client(c))) //Appengineのurlfetchを使用する
	if err != nil {
		log.Errorf(c, "Error occurred at create linebot client: %v", err)
		return
	}

	received, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			log.Warningf(c, "Linebot request status: 400")
			w.WriteHeader(400)
		} else {
			log.Warningf(c, "linebot request status: 500")
			w.WriteHeader(500)
		}
		return
	}
	for _, result := range received.Results {
		content := result.Content()
		if content != nil {
			if content.IsOperation {
				//オペレーションイベント受信
				opContent, err := content.OperationContent()
				if err != nil {
					log.Errorf(c, "Error occurred at get operation content: %v", err)
					return
				}
				sender, err2 := getSenderProfile(c, bot, opContent.Params[0])
				if err2 != nil {
					return
				}

				if content.OpType == linebot.OpTypeAddedAsFriend {
					if addFriend(c, sender.MID, sender.DisplayName) != nil {
						return
					}
					sendToAll(c, bot, sender.DisplayName+"さんが購読を開始しました。いらっしゃいませ！！")

				} else if content.OpType == linebot.OpTypeBlocked {
					if removeFriend(c, sender.MID, sender.DisplayName) != nil {
						return
					}
					sendToAll(c, bot, sender.DisplayName+"さんが購読を解除しました。さようなら！")

				} else {
					log.Warningf(c, "Unknown OpType received. OpType=%v", content.OpType)
				}

			} else if content.IsMessage && content.ContentType == linebot.ContentTypeText {
				//テキストメッセージ受信
				text, err := content.TextContent()
				if err != nil {
					log.Errorf(c, "Error occurred at parse text context: %v", err)
					return
				}
				sender, err2 := getSenderProfile(c, bot, content.From)
				if err2 != nil {
					return
				}
				sendToAll(c, bot, sender.DisplayName+"さんより\n「"+text.Text+"」")
			}
		}
	}
}

/**
 * 送信者の情報を取得する
 */
func getSenderProfile(c context.Context, bot *linebot.Client, from string) (linebot.ContactInfo, error) {
	senderProfile, err := bot.GetUserProfile([]string{from})
	if err != nil {
		log.Errorf(c, "Error occurred at get sender profile. from: %v, err: %v", from, err)
		return linebot.ContactInfo{}, err
	}
	return senderProfile.Contacts[0], nil //添字は0固定
}

/**
 * 友だち追加（データストアに購読者として登録）
 */
func addFriend(c context.Context, mid string, displayName string) error {

	//購読者を保存
	entity := subscriber{
		DisplayName: displayName,
		MID:         mid,
	}
	key := datastore.NewKey(c, "Subscriber", mid, 0, nil)
	if _, err := datastore.Put(c, key, &entity); err != nil {
		log.Errorf(c, "Error occurred at put subcriber to datastore. mid:%v, err: %v", mid, err)
		return err
	}

	//ログエントリを追加
	logEntity := logSubscriber{
		DisplayName: displayName,
		MID:         mid,
		OpType:      linebot.OpTypeAddedAsFriend,
		AddTime:     time.Now(),
	}
	logKey := datastore.NewIncompleteKey(c, "LogSubscriber", nil)
	if _, err := datastore.Put(c, logKey, &logEntity); err != nil {
		log.Errorf(c, "Error occurred at put log-subcriber to datastore. mid:%v, err: %v", mid, err)
		return err
	}
	return nil
}

/**
 * 友だち削除（データストアから削除）
 */
func removeFriend(c context.Context, mid string, displayName string) error {

	//購読者を削除
	key := datastore.NewKey(c, "Subscriber", mid, 0, nil)
	if err := datastore.Delete(c, key); err != nil {
		log.Errorf(c, "Error occurred at delete subcriber to datastore. mid:%v, err: %v", mid, err)
		return err
	}

	//ログエントリを追加
	logEntity := logSubscriber{
		DisplayName: displayName,
		MID:         mid,
		OpType:      linebot.OpTypeBlocked,
		AddTime:     time.Now(),
	}
	logKey := datastore.NewIncompleteKey(c, "LogSubscriber", nil)
	if _, err := datastore.Put(c, logKey, &logEntity); err != nil {
		log.Errorf(c, "Error occurred at put log-subcriber to datastore. mid:%v, err: %v", mid, err)
		return err
	}
	return nil
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
	mids := make([]string, len(subscribers))
	for i, current := range subscribers {
		mids[i] = current.MID
	}

	//全員に送信
	if _, err := bot.SendText(mids, message); err != nil {
		log.Errorf(c, "Error occurred at send message: %v", err)
		return err
	}
	return nil
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
