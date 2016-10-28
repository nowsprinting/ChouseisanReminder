export LINE_CHANNEL_SECRET=012345678901234567890123456789ab
export LINE_CHANNEL_ACCESS_TOKEN=u012345678901234567890123456789ab

ifdef RUN
	RUNFUNC := -run $(RUN)
endif

test:
	goapp test -v -covermode=count -coverprofile=coverage.out $(RUNFUNC)
