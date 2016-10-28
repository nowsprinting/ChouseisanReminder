package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/thingful/httpmock"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
)

/**
 * 未定義のコマンド（コマンドが空文字）
 */
func TestCommandAnalyzeEmpty(t *testing.T) {
	opt := aetest.Options{StronglyConsistentDatastore: true} //データストアに即反映
	instance, err := aetest.NewInstance(&opt)
	if err != nil {
		t.Fatalf("Failed to create aetest instance: %v", err)
	}
	defer instance.Close()

	// 評価する値
	expectedMid := "C00000000000000000000000000000000" //グループなので先頭は"C"

	// http.Requestを生成
	param := url.Values{
		"mid":        {expectedMid},
		"replyToken": {"nHuyWiB7yP5Zw52FIkcQobQuGDXCTA"},
		"text":       {""}, //空文字
	}
	req, err := instance.NewRequest("POST", "/task/analyzecommand", strings.NewReader(param.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") //必須

	// Contextとhttp.Clientは、テストコード側でインスタンス化する（モックと共通のインスタンスを使う必要があるため）
	ctx := appengine.NewContext(req)
	client := urlfetch.Client(ctx)

	// コマンド送信グループとMIDが一致する購読者エンティティを用意しておく
	entity := subscriber{}
	key := datastore.NewKey(ctx, "Subscriber", expectedMid, 0, nil)
	if _, err = datastore.Put(ctx, key, &entity); err != nil {
		t.Fatal(err)
	}

	// LINEへのReply Messageリクエストをモックする
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	actualSendMessages := []string{} //モックに送られたリプライメッセージを保持し、後で検証する
	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"POST",
			"https://api.line.me/v2/bot/message/reply",
			func(req *http.Request) (*http.Response, error) {
				defer req.Body.Close()
				if body, err := ioutil.ReadAll(req.Body); err == nil {
					actualSendMessages = append(actualSendMessages, string(body))
					return httpmock.NewStringResponse(200, "{}"), nil
				}
				return httpmock.NewStringResponse(500, "Unread post body"), nil
			},
		),
	)

	// execute
	res := httptest.NewRecorder()
	commandAnalyzeWithContext(ctx, client, res, req) //モックと同じhttp.Clientインスタンスを渡す

	// リクエストは正常終了していること
	if res.Code != http.StatusOK {
		t.Errorf("Non-expected status code: %v\n\tbody: %v", res.Code, res.Body)
	}

	// スタブがすべて呼ばれたことを検証
	if err := httpmock.AllStubsCalled(); err != nil {
		t.Errorf("Not all stubs were called: %s", err)
	}

	//送信メッセージの検証
	if !regexp.MustCompile(".*無効なコマンドです.*").MatchString(actualSendMessages[0]) {
		t.Errorf("Unmatch send message text: %v", actualSendMessages[0])
	}
}

/**
 * 調整さんハッシュ設定コマンド（正常系）
 */
func TestCommandAnalyzeSetChouseisanNormally(t *testing.T) {
	opt := aetest.Options{StronglyConsistentDatastore: true} //データストアに即反映
	instance, err := aetest.NewInstance(&opt)
	if err != nil {
		t.Fatalf("Failed to create aetest instance: %v", err)
	}
	defer instance.Close()

	// 評価する値
	expectedMid := "C00000000000000000000000000000000" //グループなので先頭は"C"
	expectedHash := "3f7ffd73ba174332ae05bd363eba8e71"

	// http.Requestを生成
	param := url.Values{
		"mid":        {expectedMid},
		"replyToken": {"nHuyWiB7yP5Zw52FIkcQobQuGDXCTA"},
		"text":       {"set chouseisan https://chouseisan.com/s?h=" + expectedHash},
	}
	req, err := instance.NewRequest("POST", "/task/analyzecommand", strings.NewReader(param.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") //必須

	// Contextとhttp.Clientは、テストコード側でインスタンス化する（モックと共通のインスタンスを使う必要があるため）
	ctx := appengine.NewContext(req)
	client := urlfetch.Client(ctx)

	// コマンド送信グループとMIDが一致する購読者エンティティを用意しておく
	entity := subscriber{}
	key := datastore.NewKey(ctx, "Subscriber", expectedMid, 0, nil)
	if _, err = datastore.Put(ctx, key, &entity); err != nil {
		t.Fatal(err)
	}

	// LINEへのReply Messageリクエストをモックする
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	actualSendMessages := []string{} //モックに送られたリプライメッセージを保持し、後で検証する
	httpmock.RegisterStubRequest(
		httpmock.NewStubRequest(
			"POST",
			"https://api.line.me/v2/bot/message/reply",
			func(req *http.Request) (*http.Response, error) {
				defer req.Body.Close()
				if body, err := ioutil.ReadAll(req.Body); err == nil {
					actualSendMessages = append(actualSendMessages, string(body))
					return httpmock.NewStringResponse(200, "{}"), nil
				}
				return httpmock.NewStringResponse(500, "Unread post body"), nil
			},
		),
	)

	// execute
	res := httptest.NewRecorder()
	commandAnalyzeWithContext(ctx, client, res, req) //モックと同じhttp.Clientインスタンスを渡す

	// リクエストは正常終了していること
	if res.Code != http.StatusOK {
		t.Errorf("Non-expected status code: %v\n\tbody: %v", res.Code, res.Body)
	}

	// スタブがすべて呼ばれたことを検証
	if err = httpmock.AllStubsCalled(); err != nil {
		t.Errorf("Not all stubs were called: %s", err)
	}

	//送信メッセージの検証
	if !regexp.MustCompile("リマインドする調整さんイベントを設定しました").MatchString(actualSendMessages[0]) {
		t.Errorf("Unmatch send message text: %v", actualSendMessages[0])
	}

	// データストアにハッシュが書き込まれていること
	actualEntity := subscriber{}
	if err = datastore.Get(ctx, key, &actualEntity); err != nil {
		t.Fatal(err)
	}
	if actualEntity.ChouseisanHash != expectedHash {
		t.Errorf("Unmatch entitiy's hash. hash='%v'", actualEntity.ChouseisanHash)
	}
}
