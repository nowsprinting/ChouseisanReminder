package main

import "testing"

/**
 * `uidtest`コマンド判定
 */
func TestIsUidtestCommand(t *testing.T) {
	type testParameter struct {
		text     string
		expected bool
	}
	testCases := []testParameter{{
		text:     "uidtest",
		expected: true,
	}, {
		text:     "    uidtest  \n\n", // 前後にノイズがあってもtrue
		expected: true,
	}, {
		text:     "uid", //コマンド誤り
		expected: false,
	}}

	for _, current := range testCases {
		actual := isUidtestCommand(current.text)
		if actual != current.expected {
			t.Errorf("Illegal return value. text:%v, returnd:%v", current.text, actual)
		}
	}
}
