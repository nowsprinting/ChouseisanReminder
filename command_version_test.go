package main

import (
	"testing"

	"google.golang.org/appengine/aetest"
)

/**
 * `version`コマンド（正常系）
 */
func TestIsVersionCommandNormally(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	text := "version"

	if isVersionCommand(c, text) == false {
		t.Error("isVersionCommand() returnd false.")
	}
}

/**
 * `version`コマンド（正常系・ノイズ付き）
 *
 * コマンド文字列の前後に、スペースや改行のノイズを入れる
 */
func TestIsVersionCommandWithNoise(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	text := "    version  \n\n"

	if isVersionCommand(c, text) == false {
		t.Error("isVersionCommand() returnd false.")
	}
}

/**
 * `version`コマンド（コマンド不一致）
 */
func TestIsVersionCommandUnmatch(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	text := "ver"

	if isVersionCommand(c, text) {
		t.Error("isVersionCommand() returnd true. But, source text is not 'version' command.")
	}
}
