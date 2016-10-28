package main

import (
	"io/ioutil"
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

func readFile(t *testing.T, filename string) string {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return string(bs)
}

/**
 * 登録イベント：グループ、ルーム（同じ扱いなのでテストはひとつ）
 */
func TestJoinGroup(t *testing.T) {
	opt := aetest.Options{StronglyConsistentDatastore: true} //データストアに即反映
	instance, err := aetest.NewInstance(&opt)
	if err != nil {
		t.Fatalf("Failed to create aetest instance: %v", err)
	}
	defer instance.Close()

	// 評価する値
	expectedMid := "C00000000000000000000000000000000" //グループなので先頭は"C"
	expectedName := expectedMid                        //グループなので同じ値
	expectedType := "join"

	// http.Requestを生成
	param := url.Values{
		"mid":        {expectedMid},
		"type":       {expectedType},
		"replyToken": {"nHuyWiB7yP5Zw52FIkcQobQuGDXCTA"},
	}
	req, err := instance.NewRequest("POST", "/task/join", strings.NewReader(param.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") //必須

	// Contextとhttp.Clientは、テストコード側でインスタンス化する（モックと共通のインスタンスを使う必要があるため）
	ctx := appengine.NewContext(req)
	client := urlfetch.Client(ctx)

	// LINEへのReply Messageリクエストをモックする
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"POST",
			"https://api.line.me/v2/bot/message/reply",
			httpmock.NewStringResponder(200, "{}"),
		),
	)

	// execute
	res := httptest.NewRecorder()
	joinWithContext(ctx, client, res, req) //モックと同じhttp.Clientインスタンスを渡す

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
	if len(subscribers) != 1 {
		t.Fatal("Subscriber entity was not put")
	}
	if subscribers[0].MID != expectedMid {
		t.Errorf("Invalid posted subscriber entity. MID='%v'", subscribers[0].MID)
	}
	if subscribers[0].DisplayName != expectedName {
		t.Errorf("Invalid posted subscriber entity. DisplayName='%v'", subscribers[0].DisplayName)
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
 * 登録イベント：ユーザ
 */
func TestJoinUser(t *testing.T) {
	opt := aetest.Options{StronglyConsistentDatastore: true} //データストアに即反映
	instance, err := aetest.NewInstance(&opt)
	if err != nil {
		t.Fatalf("Failed to create aetest instance: %v", err)
	}
	defer instance.Close()

	// 評価する値
	expectedMid := "U00000000000000000000000000000000" //ユーザなので先頭は"U"
	expectedName := "LINE taro"                        //ユーザなので、Get Profile APIで取得に行く
	expectedType := "follow"

	// http.Requestを生成
	param := url.Values{
		"mid":        {expectedMid},
		"type":       {expectedType},
		"replyToken": {"nHuyWiB7yP5Zw52FIkcQobQuGDXCTA"},
	}
	req, err := instance.NewRequest("POST", "/task/join", strings.NewReader(param.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") //必須

	// Contextとhttp.Clientは、テストコード側でインスタンス化する（モックと共通のインスタンスを使う必要があるため）
	ctx := appengine.NewContext(req)
	client := urlfetch.Client(ctx)

	// LINEへのGet ProfileおよびReply Messageリクエストをモックする
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"GET",
			"https://api.line.me/v2/bot/profile/"+expectedMid,
			httpmock.NewStringResponder(200, readFile(t, "testdata/linebot/profile.json")),
		),
	)
	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"POST",
			"https://api.line.me/v2/bot/message/reply",
			httpmock.NewStringResponder(200, "{}"),
		),
	)

	// execute
	res := httptest.NewRecorder()
	joinWithContext(ctx, client, res, req) //モックと同じhttp.Clientインスタンスを渡す

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
	if len(subscribers) != 1 {
		t.Fatal("Subscriber entity was not put")
	}
	if subscribers[0].MID != expectedMid {
		t.Errorf("Invalid posted subscriber entity. MID='%v'", subscribers[0].MID)
	}
	if subscribers[0].DisplayName != expectedName {
		t.Errorf("Invalid posted subscriber entity. DisplayName='%v'", subscribers[0].DisplayName)
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
