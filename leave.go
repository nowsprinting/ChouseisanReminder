package main

import (
	"net/http"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

/**
 * 友だち削除（データストアから削除）
 */
func leave(w http.ResponseWriter, r *http.Request) {
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
		EventType:   r.FormValue("type"),
		AddTime:     time.Now(),
	}
	logKey := datastore.NewIncompleteKey(c, "LogSubscriber", nil)
	if _, err := datastore.Put(c, logKey, &logEntity); err != nil {
		log.Errorf(c, "Error occurred at put log-subcriber to datastore. mid:%v, err: %v", mid, err)
		return
	}
}
