package main

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/line/line-bot-sdk-go/linebot"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

/**
 * コマンド実行結果をリプライ
 */
func replyMessage(c context.Context, client *http.Client, token string, message string) {
	bot, err := createBotClient(c, client)
	if err != nil {
		return
	}

	if _, err = bot.ReplyMessage(token, linebot.NewTextMessage(message)).Do(); err != nil {
		log.Errorf(c, "Error occurred at reply-message for command. err: %v", err)
	}
}

/**
 * コマンド解析
 *
 * 引数にContextとhttp.Clientを取るインナーメソッド
 */
func commandAnalyzeWithContext(c context.Context, client *http.Client, w http.ResponseWriter, r *http.Request) {
	mid := r.FormValue("mid")
	token := r.FormValue("replyToken")
	text := r.FormValue("text")

	// `set chouseisan` command
	if b, hash := isSetChouseisanCommand(c, text); b {
		if err := writeChouseisanHash(c, mid, hash); err != nil {
			message := "リマインドする調整さんイベントを設定しました"
			replyMessage(c, client, token, message)
		} else {
			message := "調整さんイベントの設定に失敗しました\n" + err.Error()
			replyMessage(c, client, token, message)
		}
		return
	}

	// Reply "invalid command" message
	message := "無効なコマンドです。\n有効なコマンドは、こちらのページをご覧ください\nhttps://" + appengine.DefaultVersionHostname(c) + "/"
	replyMessage(c, client, token, message)
}

/**
 * コマンド解析
 */
func commandAnalyze(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	commandAnalyzeWithContext(c, urlfetch.Client(c), w, r)
}
