package inkminer

import (
	"log"

	"../blockartlib"
)

func (i *InkMiner) GetBlock(hash string) (blockartlib.Block, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()

	b, ok := i.mu.blockchain[hash]
	return b, ok
}

func (i *InkMiner) mineBlock(operation blockartlib.Operation) error {
	// TODO: Jonathan - verify operation and start mining this block. Set mined block to be currentHead and create a State object

	// This maybe should be structured as "daemon" ie. an infinite for loop with
	// channels in/out so it's possible to interrupt mid block. - Tristan

	i.mu.Lock()
	defer i.mu.Unlock()

	i.mu.currentWIPBlock.Records = append(i.mu.currentWIPBlock.Records, operation)
	i.mineBlockChan <- i.mu.currentWIPBlock

	return nil
}

func (i *InkMiner) startMining() error {
	i.mineBlockChan = make(chan blockartlib.Block, 10)

	go i.minerLoop(i.mineBlockChan)

	return nil
}

// mineBlock returns the nonce, whether or not it found a valid nonce and an
// error.
func (i *InkMiner) mineWorker(block blockartlib.Block, oldNonce uint32, maxIterations int) (uint32, bool, error) {
	// TODO: spin up goroutines for mining
	return 0, false, ErrUnimplemented
}

func (i *InkMiner) minerLoop(blocks chan blockartlib.Block) {
outer:
	for {
		block := <-blocks
		nonce := uint32(0)
		found := false
		var err error
		for {
			// attempt to mine a block for a set number of iterations
			nonce, found, err = i.mineWorker(block, nonce, 1000)
			if err != nil {
				log.Printf("Mining error: %+v", err)
				continue outer
			}
			if found {
				block.Nonce = nonce
				if err := i.announceBlock(block); err != nil {
					log.Printf("Mining error: %+v", err)
					continue outer
				}

			}

			// Check if there's a new block to mine on instead.
			select {
			case block = <-blocks:
			default:
			}
		}
	}
}
