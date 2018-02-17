package inkminer

type ValidateNumWaiter struct {
	done        chan string
	err         chan error
	validateNum uint8
}
