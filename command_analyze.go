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
 * コマンド解析
 *
 * 引数にContextとhttp.Clientを取るインナーメソッド
 */
func commandAnalyzeWithContext(c context.Context, client *http.Client, w http.ResponseWriter, r *http.Request) {
	bot, err := createBotClient(c, client)
	if err != nil {
		return
	}

	token := r.FormValue("replyToken")
	mid := r.FormValue("mid")

	// Reply message
	message := "無効なコマンドです。\n有効なコマンドは、こちらのページをご覧ください\nhttps://" + appengine.DefaultVersionHostname(c) + "/"
	if _, err = bot.ReplyMessage(token, linebot.NewTextMessage(message)).Do(); err != nil {
		log.Errorf(c, "Error occurred at reply-message for command. mid:%v, err: %v", mid, err)
	}
}

/**
 * コマンド解析
 */
func commandAnalyze(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	commandAnalyzeWithContext(c, urlfetch.Client(c), w, r)
}
