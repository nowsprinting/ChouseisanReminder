package main

import (
	"regexp"

	"golang.org/x/net/context"
)

/**
 * `version`コマンドであればtrueを返す
 */
func isVersionCommand(c context.Context, command string) bool {
	pattern := regexp.MustCompile(`^[ \n]*version[ \n]*$`)
	return pattern.MatchString(command)
}
