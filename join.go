package main

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/line/line-bot-sdk-go/linebot"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

/**
 * 送信者の表示名を取得する
 * ユーザしか取得できないので、ルームおよびグループではidをそのまま返す
 */
func getSenderName(c context.Context, bot *linebot.Client, from string) string {
	if len(from) == 0 {
		log.Warningf(c, "Parameter `mid` was not specified.")
		return from
	}
	if from[0:1] == "U" {
		senderProfile, err := bot.GetProfile(from).Do()
		if err != nil {
			log.Warningf(c, "Error occurred at get sender profile. from: %v, err: %v", from, err)
			return from
		}
		return senderProfile.DisplayName
	}
	return from
}

/**
 * 友だち追加（データストアに購読者として登録）
 *
 * 引数にContextとhttp.Clientを取るインナーメソッド
 */
func joinWithContext(c context.Context, client *http.Client, w http.ResponseWriter, r *http.Request) {
	bot, err := createBotClient(c, client)
	if err != nil {
		return
	}

	//購読者プロファイルを取得
	mid := r.FormValue("mid")
	senderName := getSenderName(c, bot, mid)

	//購読者を保存（リマインドタイミングのデフォルトは3日前の8:00）
	entity := subscriber{
		DisplayName:    senderName,
		MID:            mid,
		ChouseisanHash: "",
		RemindBefore:   3,
		RemindTime:     8,
	}
	key := datastore.NewKey(c, "Subscriber", mid, 0, nil)
	if _, err = datastore.Put(c, key, &entity); err != nil {
		log.Errorf(c, "Error occurred at put subcriber to datastore. mid:%v, err: %v", mid, err)
		return
	}

	//ログエントリを追加
	logEntity := logSubscriber{
		DisplayName: senderName,
		MID:         mid,
		EventType:   r.FormValue("type"),
		AddTime:     time.Now(),
	}
	logKey := datastore.NewIncompleteKey(c, "LogSubscriber", nil)
	if _, err = datastore.Put(c, logKey, &logEntity); err != nil {
		log.Errorf(c, "Error occurred at put log-subcriber to datastore. mid:%v, err: %v", mid, err)
		return
	}

	// Reply message
	message := "リマインダを登録しました！\n使いかたはこちらのページをご覧ください\nhttps://" + appengine.DefaultVersionHostname(c) + "/"
	if _, err = bot.ReplyMessage(r.FormValue("replyToken"), linebot.NewTextMessage(message)).Do(); err != nil {
		log.Errorf(c, "Error occurred at reply-message for follow/join. mid:%v, err: %v", mid, err)
	}
}

/**
 * 友だち追加（データストアに購読者として登録）
 */
func join(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	joinWithContext(c, urlfetch.Client(c), w, r)
}
