export LINE_CHANNEL_SECRET=012345678901234567890123456789ab
export LINE_CHANNEL_ACCESS_TOKEN=u012345678901234567890123456789ab

ifdef RUN
	RUNFUNC := -run $(RUN)
endif

version:
	echo package main > version.go
	echo const version = \"$(shell git describe --tags)\" >> version.go

test: version
		goapp test -v -covermode=count -coverprofile=coverage.out $(RUNFUNC)

deploy: version
	goapp deploy
