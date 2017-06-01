package main

import "regexp"

/**
 * `uidtest`コマンドであればtrueを返す
 */
func isUidtestCommand(command string) bool {
	pattern := regexp.MustCompile(`^[ \n]*uidtest[ \n]*$`)
	return pattern.MatchString(command)
}
