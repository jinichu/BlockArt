
.PHONY: test
test:
	go test -timeout 30s ./inkminer ./server ./blockartlib ./crypto ./integration ./stopper ./serverold
