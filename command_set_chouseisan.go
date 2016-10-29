package main

import (
	"regexp"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"golang.org/x/net/context"
)

/**
 * `set chouseisan`コマンドであれば、調整さんハッシュを返す
 */
func isSetChouseisanCommand(command string) (bool, string) {
	pattern := regexp.MustCompile(`^[ \n]*set chouseisan https:\/\/chouseisan\.com\/s\?h=(\w+)[ \n]*$`)
	matches := pattern.FindStringSubmatch(command)
	if len(matches) == 2 {
		return true, matches[1]
	}
	return false, ""
}

/**
 * 購読者エンティティに、調整さんハッシュを書き込む
 */
func writeChouseisanHash(c context.Context, mid string, hash string) error {
	var entity subscriber

	key := datastore.NewKey(c, "Subscriber", mid, 0, nil)
	if err := datastore.Get(c, key, &entity); err != nil {
		log.Errorf(c, "Error occurred at get Subscriber entity. mid:%v err:%v", mid, err)
		return err
	}

	entity.ChouseisanHash = hash
	if _, err := datastore.Put(c, key, &entity); err != nil {
		log.Errorf(c, "Error occurred at put Subscriber entity. mid:%v err:%v", mid, err)
		return err
	}
	return nil
}
