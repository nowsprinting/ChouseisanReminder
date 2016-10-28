package main

import (
	"regexp"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

/**
 * `set name`コマンドであれば、指定された名前部分を返す
 */
func isSetNameCommand(c context.Context, command string) (bool, string) {
	pattern := regexp.MustCompile(`^[ \n]*set name (.+)[ \n]*$`)
	matches := pattern.FindStringSubmatch(command)
	if len(matches) == 2 {
		return true, matches[1]
	}
	return false, ""
}

/**
 * 購読者エンティティに、グループ名を書き込む
 */
func writeName(c context.Context, mid string, name string) error {
	var entity subscriber

	key := datastore.NewKey(c, "Subscriber", mid, 0, nil)
	if err := datastore.Get(c, key, &entity); err != nil {
		log.Errorf(c, "Error occurred at get Subscriber entity. mid:%v err:%v", mid, err)
		return err
	}

	entity.DisplayName = name
	if _, err := datastore.Put(c, key, &entity); err != nil {
		log.Errorf(c, "Error occurred at put Subscriber entity. mid:%v err:%v", mid, err)
		return err
	}
	return nil
}
