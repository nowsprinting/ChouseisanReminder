package main

import (
	"net/http"
	"strconv"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

/**
 * スタンプを送信（送信者名を入れたテキストメッセージ＋スタンプ）
 */
func shareSticker(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	bot, err := createBotClient(c)
	if err != nil {
		return
	}
	mid := r.FormValue("mid")
	stkid, err := strconv.Atoi(r.FormValue("stkid"))
	if err != nil {
		log.Errorf(c, "Error occurred at parse stkid. err: %v", err)
		return
	}
	stkpkgid, err := strconv.Atoi(r.FormValue("stkpkgid"))
	if err != nil {
		log.Errorf(c, "Error occurred at parse stkpkgid. err: %v", err)
		return
	}
	stkver, err := strconv.Atoi(r.FormValue("stkver"))
	if err != nil {
		log.Errorf(c, "Error occurred at parse stkver. err: %v", err)
		return
	}

	//データストアから購読者のMIDを取得
	q := datastore.NewQuery("Subscriber")
	var subscribers []subscriber
	if _, err = q.GetAll(c, &subscribers); err != nil {
		log.Errorf(c, "Error occurred at get-all from datastore. err: %v", err)
		return
	}
	mids := make([]string, len(subscribers))
	for i, current := range subscribers {
		mids[i] = current.MID
	}

	//送信者名を全員にブロードキャスト
	sender, err := getSenderProfile(c, bot, mid)
	if err != nil {
		log.Errorf(c, "Error occurred at get sender profile: %v", err)
		return
	}
	if _, err := bot.SendText(mids, sender.DisplayName+"さんより"); err != nil {
		log.Errorf(c, "Error occurred at send message: %v", err)
		return
	}

	//全員にスタンプを送信
	if _, err := bot.SendSticker(mids, stkid, stkpkgid, stkver); err != nil {
		log.Errorf(c, "Error occurred at send message: %v", err)
		return
	}
}
