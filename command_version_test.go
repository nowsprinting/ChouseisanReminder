package main

import "testing"

/**
 * `version`コマンド判定
 */
func TestIsVersionCommand(t *testing.T) {
	type testParameter struct {
		text     string
		expected bool
	}
	testCases := []testParameter{{
		text:     "version",
		expected: true,
	}, {
		text:     "    version  \n\n", // 前後にノイズがあってもtrue
		expected: true,
	}, {
		text:     "ver", //コマンド誤り
		expected: false,
	}}

	for _, current := range testCases {
		actual := isVersionCommand(current.text)
		if actual != current.expected {
			t.Errorf("Illegal return value. text:%v, returnd:%v", current.text, actual)
		}
	}
}
