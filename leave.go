package main

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

/**
 * 友だち削除（データストアから削除）
 *
 * 引数にContextとhttp.Clientを取るインナーメソッド
 */
func leaveWithContext(c context.Context, client *http.Client, w http.ResponseWriter, r *http.Request) {
	var entity subscriber
	mid := r.FormValue("mid")

	//購読者エンティティを抽出
	key := datastore.NewKey(c, "Subscriber", mid, 0, nil)
	if err := datastore.Get(c, key, &entity); err != nil {
		log.Errorf(c, "Error occurred at get Subscriber entity. mid:%v err: %v", mid, err)
		return
	}

	//購読者を削除
	if err := datastore.Delete(c, key); err != nil {
		log.Errorf(c, "Error occurred at delete subcriber to datastore. mid:%v, err: %v", mid, err)
		return
	}

	//ログエントリを追加
	logEntity := logSubscriber{
		DisplayName: entity.DisplayName,
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

/**
 * 友だち削除（データストアから削除）
 */
func leave(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	leaveWithContext(c, urlfetch.Client(c), w, r)
}
