
.PHONY: test
test:
	go test -timeout 30s ./inkminer ./server ./blockartlib ./crypto ./integration ./stopper ./serverold

.PHONY: loc
loc:
	cloc --3 --exclude-dir=bower_components .
