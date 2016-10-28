package main

import (
	"testing"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

/**
* `set chouseisan`コマンド判定とハッシュの取り出し（正常系）
 */
func TestIsSetChouseisanCommandNormally(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	expectedHash := "3f7ffd73ba174332ae05bd363eba8e71"
	text := "set chouseisan https://chouseisan.com/s?h=" + expectedHash

	if b, hash := isSetChouseisanCommand(c, text); b == false {
		t.Errorf("isSetChouseisanCommand() returnd false.")
	} else if hash != expectedHash {
		t.Errorf("Unmatch chouseisan hash. hash:%v", hash)
	}
}

/**
* `set chouseisan`コマンド判定とハッシュの取り出し（コマンド不一致）
 */
func TestIsSetChouseisanCommandUnmatch(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	text := "set hash https://chouseisan.com/s?h=3f7ffd73ba174332ae05bd363eba8e71"

	if b, _ := isSetChouseisanCommand(c, text); b {
		t.Error("isSetChouseisanCommand() returnd true. But, source text is not 'set chouseisan' command.")
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
