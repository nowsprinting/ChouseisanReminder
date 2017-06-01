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
	if b, hash := isSetChouseisanCommand(text); b {
		if err := writeChouseisanHash(c, mid, hash); err != nil {
			message := "調整さんイベントの設定に失敗しました\n" + err.Error()
			replyMessage(c, client, token, message)
		} else {
			message := "リマインドする調整さんイベントを設定しました"
			replyMessage(c, client, token, message)
		}
		return
	}

	// `set name` command
	if b, name := isSetNameCommand(text); b {
		if err := writeName(c, mid, name); err != nil {
			message := "グループ（もしくはトークルーム）の名前の設定に失敗しました\n" + err.Error()
			replyMessage(c, client, token, message)
		} else {
			message := "グループ（もしくはトークルーム）の名前を設定しました"
			replyMessage(c, client, token, message)
		}
		return
	}

	// `uidtest` command（user idを取得してユーザネームをレスポンスする）
	if isUidtestCommand(text) {
		bot, err := createBotClient(c, client)
		if err != nil {
			return
		}
		uid := r.FormValue("uid")
		message := ""
		if len(uid) > 0 && uid[0:1] == "U" {
			senderProfile, err := bot.GetProfile(uid).Do()
			if err != nil {
				log.Warningf(c, "Error occurred at get sender profile. uid: %v, err: %v", uid, err)
				message = "userProfile取得失敗(" + uid + ")"
			} else {
				message = "今のメッセージ送信者は、" + senderProfile.DisplayName + "さんです"
			}
		} else {
			message = "userId取得失敗(" + uid + ")"
		}
		replyMessage(c, client, token, message)
	}

	// `version` command
	if isVersionCommand(text) {
		message := version
		replyMessage(c, client, token, message)
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
