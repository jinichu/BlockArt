
.PHONY: test
test:
	go test -v ./inkminer ./server ./blockartlib ./crypto ./integration ./stopper
