package inkminer

type ValidateNumWaiter struct {
	done        chan string
	validateNum uint8
}
