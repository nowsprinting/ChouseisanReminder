package main

import "regexp"

/**
 * `version`コマンドであればtrueを返す
 */
func isVersionCommand(command string) bool {
	pattern := regexp.MustCompile(`^[ \n]*version[ \n]*$`)
	return pattern.MatchString(command)
}
