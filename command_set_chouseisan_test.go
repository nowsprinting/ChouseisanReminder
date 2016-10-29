package main

import (
	"testing"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

/**
 * `set chouseisan`コマンド判定とハッシュの取り出し
 */
func TestIsSetChouseisanCommandNormally(t *testing.T) {
	type testParameter struct {
		text         string
		expectedIs   bool
		expectedHash string
	}
	testCases := []testParameter{{
		text:         "set chouseisan https://chouseisan.com/s?h=3f7ffd73ba174332ae05bd363eba8e71",
		expectedIs:   true,
		expectedHash: "3f7ffd73ba174332ae05bd363eba8e71",
	}, {
		text:         "  set chouseisan https://chouseisan.com/s?h=3f7ffd73ba174332ae05bd363eba8e71\n\n", // 前後にノイズがあってもtrue
		expectedIs:   true,
		expectedHash: "3f7ffd73ba174332ae05bd363eba8e71",
	}, {
		text:         "set hash https://chouseisan.com/s?h=3f7ffd73ba174332ae05bd363eba8e71", // コマンド誤り
		expectedIs:   false,
		expectedHash: "",
	}}

	for _, current := range testCases {
		actualIs, actualHash := isSetChouseisanCommand(current.text)
		if actualIs != current.expectedIs {
			t.Errorf("Illegal return value. text:%v, returnd:%v", current.text, actualIs)
		}
		if actualHash != current.expectedHash {
			t.Errorf("Illegal return value. text:%v, returnd:%v", current.text, actualHash)
		}
	}
}

/**
 * データストアに調整さんハッシュを書き込む関数のテスト（正常系）
 */
func TestWriteChouseisanHashNormally(t *testing.T) {
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

	expectedHash := "3f7ffd73ba174332ae05bd363eba8e71"
	mid := "C00000000000000000000000000000000"

	// 更新される購読者エンティティを用意しておく
	entity := subscriber{
		MID: mid,
	}
	key := datastore.NewKey(c, "Subscriber", mid, 0, nil)
	if _, err = datastore.Put(c, key, &entity); err != nil {
		t.Fatal(err)
	}

	// execute
	if err := writeChouseisanHash(c, mid, expectedHash); err != nil {
		t.Fatal(err)
	}

	// データストアにハッシュが書き込まれていること
	var actualEntity subscriber
	if err = datastore.Get(c, key, &actualEntity); err != nil {
		t.Fatal(err)
	}
	if actualEntity.ChouseisanHash != expectedHash {
		t.Errorf("Unmatch entitiy's hash. hash='%v'", actualEntity.ChouseisanHash)
	}
}
