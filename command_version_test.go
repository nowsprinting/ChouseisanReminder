package main

import "testing"

/**
 * `version`コマンド（正常系）
 */
func TestIsVersionCommandNormally(t *testing.T) {
	text := "version"

	if isVersionCommand(text) == false {
		t.Error("isVersionCommand() returnd false.")
	}
}

/**
 * `version`コマンド（正常系・ノイズ付き）
 *
 * コマンド文字列の前後に、スペースや改行のノイズを入れる
 */
func TestIsVersionCommandWithNoise(t *testing.T) {
	text := "    version  \n\n"

	if isVersionCommand(text) == false {
		t.Error("isVersionCommand() returnd false.")
	}
}

/**
 * `version`コマンド（コマンド不一致）
 */
func TestIsVersionCommandUnmatch(t *testing.T) {
	text := "ver"

	if isVersionCommand(text) {
		t.Error("isVersionCommand() returnd true. But, source text is not 'version' command.")
	}
}
