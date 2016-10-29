package main

import (
	"testing"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

/**
 * `set name`コマンド判定と指定された名前の取り出し（正常系）
 */
func TestIsSetNameCommandNormally(t *testing.T) {
	expectedName := "テストグループ"
	text := "set name " + expectedName

	if b, name := isSetNameCommand(text); b == false {
		t.Errorf("isSetNameCommand() returnd false.")
	} else if name != expectedName {
		t.Errorf("Unmatch group name. name:%v", name)
	}
}

/**
 * `set name`コマンド判定と指定された名前の取り出し（正常系・ノイズ付き）
 *
 * コマンド文字列の前後に、スペースや改行のノイズを入れる
 */
func TestIsSetNameCommandWithNoise(t *testing.T) {
	expectedName := "テストグループ"
	text := "   set name " + expectedName + "\n"

	if b, name := isSetNameCommand(text); b == false {
		t.Errorf("isSetNameCommand() returnd false.")
	} else if name != expectedName {
		t.Errorf("Unmatch group name. name:%v", name)
	}
}

/**
 * `set name`コマンド判定と指定された名前の取り出し（コマンド不一致）
 */
func TestIsSetNameCommandUnmatch(t *testing.T) {
	text := "set your name コマンド間違ったグループ"

	if b, _ := isSetNameCommand(text); b {
		t.Error("isSetNameCommand() returnd true. But, source text is not 'set name' command.")
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
