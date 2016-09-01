#!/bin/sh

# define environment variables
export CHOUSEISAN_EVENT_HASH=3f7ffd73ba174332ae05bd363eba8e71
export LINE_CHANNEL_ID=9999999999
export LINE_CHANNEL_SECRET=012345678901234567890123456789ab
export LINE_CHANNEL_MID=u012345678901234567890123456789ab

# run test
goapp test -v -covermode=count -coverprofile=coverage.out
