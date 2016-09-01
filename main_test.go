package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"

	"google.golang.org/appengine/aetest"
)

/**
 * LINEからのコールバックリクエストを生成する
 */
func createRequest(instance aetest.Instance, url string, file string) (*http.Request, error) {

	// リクエストボディをファイルから読み込み
	body, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	byteBody, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	// http.Requestを生成
	req, err := instance.NewRequest("POST", url, bytes.NewBuffer(byteBody))
	if err != nil {
		return nil, err
	}

	// ヘッダに署名を付与
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	hash := hmac.New(sha256.New, []byte(channelSecret))
	hash.Write(byteBody)
	encoded := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	req.Header.Add("X-LINE-ChannelSignature", encoded)

	return req, nil
}

/**
 * 友だち登録のコールバック処理
 *
 * LINEからのコールバックをパースして、Task Queueに登録するところまでを確認
 */
func TestLineCallbackAddfriend(t *testing.T) {
	instance, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	// Task Queueの起動をhttpmockで捕捉する
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST",
		"/task/addfriend",
		httpmock.NewStringResponder(200, ""),
	)

	// リクエストの組み立て
	req, err := createRequest(instance, "/line/callback", "testdata/line/add_friend.json")
	if err != nil {
		t.Fatal(err)
	}

	// execute
	res := httptest.NewRecorder()
	lineCallback(res, req)

	// LINEには200が返ること
	if res.Code != http.StatusOK {
		t.Fatalf("Non-expected status code: %v\n\tbody: %v", res.Code, res.Body)
	}

	//TODO: taskqueueに追加されたことを確認するべき。httpmock側で補足できる？
	//TODO: httpmockに渡されたmid文字列を確認できる？
}
