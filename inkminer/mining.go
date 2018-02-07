package inkminer

import (
	"log"

	"../blockartlib"
	//"crypto/cipher"
	"image"
)

func (i *InkMiner) GetBlock(hash string) (blockartlib.Block, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()

	b, ok := i.mu.blockchain[hash]
	return b, ok
}

// Consumes an operation and attempts
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
		block := <-blocks 			// Grab a block from the channel
		nonce := uint32(0) 			//
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

// Given a particular blockHash, generate a new state by walking through the blockchain
// Automatically adds states if they do not exist already
// TODO: Complete this
func (i *InkMiner) CalculateState(blockHash string) (newState State, err error){
	newState = State{}
	newState.inkLevels = make(map[string]uint32)
	newState.shapeOwners = make(map[string]string)
	newState.shapes = make(map[string]string)

	err = nil

	i.mu.Lock()
	defer i.mu.Unlock()

	// Check if the block exists on the blockchain, fail if not found
	block, ok := i.mu.blockchain[blockHash]
	if !ok {
		return newState, blockartlib.InvalidBlockHashError(blockHash)
	}

	if block.PrevBlock == "" {
		// Block is a genesis block; the state is blank
		return newState, nil
	}

	// If the state was already previously calculated, simply return it
	if state, ok := i.states[blockHash]; ok {
		return state, nil
	}

	// Begin walking through the blockchain, adding every non-calculated block to a worklist to compute
	workListStack := make([]blockartlib.Block, 0)
	workListStack = append(workListStack, block)

	lastState := State{}
	foundState := false
	// Keep walking until we hit a genesis block, or a pre-computed block
	for err == nil && block.PrevBlock != "" {
		// Grab the next block on the chain
		nextHash := block.PrevBlock
		block, ok := i.mu.blockchain[nextHash]
		// If we didn't find the next block, something went wrong
		if !ok {
			err = blockartlib.InvalidBlockHashError(nextHash)
			return newState, err
		}

		// Check if the next block has a state associated with it
		nextHash, err := block.Hash()
		if err != nil {
			return newState, err
		}
		_, ok = i.states[nextHash]
		if ok {
			// We did it! We found a block that has some kind of state attached to it!
			// Save this state, we will use it to compute the rest of our worklist
			foundState = true
			lastState = i.states[nextHash]
			break
		} else {
			// Not found just yet, add this block to the worklist and keep going
			workListStack = append(workListStack, block)
			continue
		}
	}

	// Check if we did not find a block to work with
	if !foundState {
		lastState = State{}
		lastState.shapes = make(map[string]string)
		lastState.shapeOwners = make(map[string]string)
		lastState.inkLevels = make(map[string]uint32)
	}

	// Now, attempt to work through the worklist
	for i := len(workListStack) - 1; i > 0; i-- {
		workingBlock := workListStack[i]
		createdState := foundState
		for _, op := range workingBlock.Records {
			
		}
	}

	//// If the block is not a genesis block, and no errors were encountered
	//for (block.PrevBlock != "") && (err == nil) {
	//	//
	//
	//	block, ok := i.mu.blockchain[block.PrevBlock]
	//	if !ok {
	//		return newState, blockartlib.InvalidBlockHashError(block.PrevBlock)
	//	}
	//
	//	workListStack = append(workListStack, block)
	//}

	return newState, err
}

func (i *InkMiner) retrieveState(blockHash string) {

}

// Starts from the
func (i *InkMiner) autoAddStates() {

}
