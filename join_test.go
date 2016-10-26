package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

/**
 * 登録イベント：グループ
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

	// LINEへのReply Messageリクエストをモックする
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST",
		"https://api.line.me/v2/bot/message/reply",
		httpmock.NewStringResponder(200, "{}"),
	)

	// http.Requestを生成
	param := url.Values{
		"mid":        {expectedMid},
		"type":       {"join"},
		"replyToken": {"nHuyWiB7yP5Zw52FIkcQobQuGDXCTA"},
	}
	req, err := instance.NewRequest("POST", "/task/join", strings.NewReader(param.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	ctx := appengine.NewContext(req)

	// execute
	res := httptest.NewRecorder()
	join(res, req)

	// リクエストは正常終了していること
	if res.Code != http.StatusOK {
		t.Fatalf("Non-expected status code: %v\n\tbody: %v", res.Code, res.Body)
	}

	// データストアの内容を確認
	subscribers := []subscriber{}
	_, err = datastore.NewQuery("Subscriber").GetAll(ctx, &subscribers)
	if err != nil {
		t.Fatalf("Failed to verify datastore: %v", err)
	}
	if len(subscribers) != 1 {
		t.Fatal("Subscriber entity was not put")
	}
	if subscribers[0].MID != expectedMid {
		t.Fatalf("Invalid posted subscriber entity. MID=%v", subscribers[0].MID)
	}
	if subscribers[0].DisplayName != expectedName {
		t.Fatalf("Invalid posted subscriber entity. DisplayName=%v", subscribers[0].DisplayName)
	}
}

/**
 * 登録イベント：ルーム
 */
func TestJoinRoom(t *testing.T) {
	t.Fail()
}

/**
 * 登録イベント：ユーザ
 */
func TestJoinUser(t *testing.T) {
	t.Fail()
}
