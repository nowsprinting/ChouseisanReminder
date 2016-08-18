package main

import (
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

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
	bot, err := linebot.NewClient(channelID, channelSecret, channelMID,
		linebot.WithHTTPClient(urlfetch.Client(c))) //Appengineのurlfetchを使用する
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
		if content != nil && content.IsMessage && content.ContentType == linebot.ContentTypeText {
			text, err := content.TextContent()

			//送信者のディスプレイネームを取得
			user, err := bot.GetUserProfile([]string{content.From})

			//メッセージ送信
			_, err = bot.SendText([]string{content.From},
				user.Contacts[0].DisplayName+"さんより\n「"+text.Text+"」")
			if err != nil {
				log.Errorf(c, "Error occurred at send text:%v", err)
			}
		}
	}
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
