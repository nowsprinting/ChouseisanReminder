package main

import (
	"testing"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

/**
 * `set name`コマンド判定と指定された名前の取り出し
 */
func TestIsSetNameCommand(t *testing.T) {
	type testParameter struct {
		text         string
		expectedIs   bool
		expectedName string
	}
	testCases := []testParameter{{
		text:         "set name テストグループ",
		expectedIs:   true,
		expectedName: "テストグループ",
	}, {
		text:         "     set name テストグループ\n\n", // 前後にノイズがあってもtrue
		expectedIs:   true,
		expectedName: "テストグループ",
	}, {
		text:         "set your name コマンド間違ったグループ", // コマンド誤り
		expectedIs:   false,
		expectedName: "",
	}}

	for _, current := range testCases {
		actualIs, actualName := isSetNameCommand(current.text)
		if actualIs != current.expectedIs {
			t.Errorf("Illegal return value. text:%v, returnd:%v", current.text, actualIs)
		}
		if actualName != current.expectedName {
			t.Errorf("Illegal return value. text:%v, returnd:%v", current.text, actualName)
		}
	}
}

/**
 * データストアにグループ名を書き込む関数のテスト（正常系）
 */
func TestWriteNameNormally(t *testing.T) {
	opt := aetest.Options{StronglyConsistentDatastore: true} //データストアに即反映
	instance, err := aetest.NewInstance(&opt)
	if err != nil {
		t.Fatalf("Failed to create aetest instance: %v", err)
	}
	defer instance.Close()

	// Contextが必要なので、ダミーのhttp.Request
	req, err := instance.NewRequest("POST", "/task/analyzecommand", nil)
	if err != nil {
		t.Fatal(err)
	}
	c := appengine.NewContext(req)

	mid := "C00000000000000000000000000000000"
	expectedName := "テストグループ"

	// 更新される購読者エンティティを用意しておく
	entity := subscriber{
		MID:         mid,
		DisplayName: "更新前のグループ名",
	}
	key := datastore.NewKey(c, "Subscriber", mid, 0, nil)
	if _, err = datastore.Put(c, key, &entity); err != nil {
		t.Fatal(err)
	}

	// execute
	if err := writeName(c, mid, expectedName); err != nil {
		t.Fatal(err)
	}

	// データストアにグループ名が書き込まれていること
	var actualEntity subscriber
	if err = datastore.Get(c, key, &actualEntity); err != nil {
		t.Fatal(err)
	}
	if actualEntity.DisplayName != expectedName {
		t.Errorf("Unmatch entitiy's DisplayName. DisplayName='%v'", actualEntity.DisplayName)
	}
}
