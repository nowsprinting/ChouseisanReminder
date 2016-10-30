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
	req.Header.Add("X-LINE-Signature", encoded)

	return req, nil
}

/**
 * LINEからのWebhookパース処理のテスト。Text Messageに対して、何もアクションしないケース
 *
 * 本来、joinイベントなどでTask Queueを起動するところまでを確認したいが、検証手段がないため、このテストのみ。
 */
func TestLineCallbackDoNotiong(t *testing.T) {
	instance, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer instance.Close()

	// リクエストの組み立て
	req, err := createRequest(instance, "/line/callback", "testdata/linebot/webhook_text_message.json")
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
}
