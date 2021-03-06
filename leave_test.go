package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/thingful/httpmock"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
)

/**
 * 削除イベント：グループ、ルーム（同じ扱いなのでテストはひとつ）
 */
func TestLeaveGroup(t *testing.T) {
	opt := aetest.Options{StronglyConsistentDatastore: true} //データストアに即反映
	instance, err := aetest.NewInstance(&opt)
	if err != nil {
		t.Fatalf("Failed to create aetest instance: %v", err)
	}
	defer instance.Close()

	// 評価する値
	expectedMid := "C00000000000000000000000000000000" //グループなので先頭は"C"
	expectedName := "テストグループ"
	expectedType := "leave"

	// http.Requestを生成
	param := url.Values{
		"mid":  {expectedMid},
		"type": {expectedType},
	}
	req, err := instance.NewRequest("POST", "/task/leave", strings.NewReader(param.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") //必須

	// Contextとhttp.Clientは、テストコード側でインスタンス化する（モックと共通のインスタンスを使う必要があるため）
	ctx := appengine.NewContext(req)
	client := urlfetch.Client(ctx)

	// 削除される購読者エンティティを用意しておく
	entity := subscriber{
		MID:         expectedMid,
		DisplayName: expectedName,
	}
	key := datastore.NewKey(ctx, "Subscriber", expectedMid, 0, nil)
	if _, err = datastore.Put(ctx, key, &entity); err != nil {
		t.Fatal(err)
	}

	// execute
	res := httptest.NewRecorder()
	leaveWithContext(ctx, client, res, req) //モックと同じhttp.Clientインスタンスを渡す

	// リクエストは正常終了していること
	if res.Code != http.StatusOK {
		t.Errorf("Non-expected status code: %v\n\tbody: %v", res.Code, res.Body)
	}

	// データストアの内容を確認（購読者エンティティ）
	subscribers := []subscriber{}
	_, err = datastore.NewQuery("Subscriber").GetAll(ctx, &subscribers)
	if err != nil {
		t.Fatal(err)
	}
	if len(subscribers) != 0 {
		t.Fatal("Subscriber entity is not remove")
	}

	// データストアの内容を確認（ログエンティティ）
	logSubscribers := []logSubscriber{}
	_, err = datastore.NewQuery("LogSubscriber").GetAll(ctx, &logSubscribers)
	if err != nil {
		t.Fatal(err)
	}
	if len(logSubscribers) != 1 {
		t.Fatal("LogSubscriber entity was not put")
	}
	if logSubscribers[0].MID != expectedMid {
		t.Errorf("Invalid posted LogSubscriber entity. MID='%v'", logSubscribers[0].MID)
	}
	if logSubscribers[0].DisplayName != expectedName {
		t.Errorf("Invalid posted LogSubscriber entity. DisplayName='%v'", logSubscribers[0].DisplayName)
	}
	if logSubscribers[0].EventType != expectedType {
		t.Errorf("Invalid posted LogSubscriber entity. EventType='%v'", logSubscribers[0].EventType)
	}
}

/**
 * 削除イベント：ユーザ
 */
func TestLeaveUser(t *testing.T) {
	opt := aetest.Options{StronglyConsistentDatastore: true} //データストアに即反映
	instance, err := aetest.NewInstance(&opt)
	if err != nil {
		t.Fatalf("Failed to create aetest instance: %v", err)
	}
	defer instance.Close()

	// 評価する値
	expectedMid := "U00000000000000000000000000000000" //ユーザなので先頭は"U"
	expectedName := "LINE taro"
	expectedType := "unfollow"

	// http.Requestを生成
	param := url.Values{
		"mid":  {expectedMid},
		"type": {expectedType},
	}
	req, err := instance.NewRequest("POST", "/task/leave", strings.NewReader(param.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") //必須

	// Contextとhttp.Clientは、テストコード側でインスタンス化する（モックと共通のインスタンスを使う必要があるため）
	ctx := appengine.NewContext(req)
	client := urlfetch.Client(ctx)

	// 外部アクセスはないはずだが、抑止のためにhttpmockを有効化
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	// 削除される購読者エンティティを用意しておく
	entity := subscriber{
		MID:         expectedMid,
		DisplayName: expectedName,
	}
	key := datastore.NewKey(ctx, "Subscriber", expectedMid, 0, nil)
	if _, err = datastore.Put(ctx, key, &entity); err != nil {
		t.Fatal(err)
	}

	// execute
	res := httptest.NewRecorder()
	leaveWithContext(ctx, client, res, req) //モックと同じhttp.Clientインスタンスを渡す

	// リクエストは正常終了していること
	if res.Code != http.StatusOK {
		t.Errorf("Non-expected status code: %v\n\tbody: %v", res.Code, res.Body)
	}

	// スタブがすべて呼ばれたことを検証
	if err = httpmock.AllStubsCalled(); err != nil {
		t.Errorf("Not all stubs were called: %s", err)
	}

	// データストアの内容を確認（購読者エンティティ）
	subscribers := []subscriber{}
	_, err = datastore.NewQuery("Subscriber").GetAll(ctx, &subscribers)
	if err != nil {
		t.Fatal(err)
	}
	if len(subscribers) != 0 {
		t.Errorf("Subscriber entity is not remove")
	}

	// データストアの内容を確認（ログエンティティ）
	logSubscribers := []logSubscriber{}
	_, err = datastore.NewQuery("LogSubscriber").GetAll(ctx, &logSubscribers)
	if err != nil {
		t.Fatal(err)
	}
	if len(logSubscribers) != 1 {
		t.Fatal("LogSubscriber entity was not put")
	}
	if logSubscribers[0].MID != expectedMid {
		t.Errorf("Invalid posted LogSubscriber entity. MID='%v'", logSubscribers[0].MID)
	}
	if logSubscribers[0].DisplayName != expectedName {
		t.Errorf("Invalid posted LogSubscriber entity. DisplayName='%v'", logSubscribers[0].DisplayName)
	}
	if logSubscribers[0].EventType != expectedType {
		t.Errorf("Invalid posted LogSubscriber entity. EventType='%v'", logSubscribers[0].EventType)
	}
}
