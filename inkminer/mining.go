package inkminer

import (
	"log"

	"../blockartlib"
	//"crypto/cipher"
	"fmt"
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
	difficulty := i.settings.PoWDifficultyOpBlock
	if len(block.Records) == 0 {
		difficulty = i.settings.PoWDifficultyNoOpBlock
	}

	for i := 0; i < maxIterations; i++ {
		oldNonce += 1
		block.Nonce = oldNonce
		hash, err := block.Hash()
		if err != nil {
			return 0, false, err
		}
		if uint8(numZeros(hash)) == difficulty {
			return oldNonce, true, nil
		}
	}
	return oldNonce, false, nil
}

func numZeros(str string) int {
	for i, c := range str {
		if c != '0' {
			return i
		}
	}
	return len(str)
}

func (i *InkMiner) minerLoop(blocks chan blockartlib.Block) {
outer:
	for {
		block := <-blocks  // Grab a block from the channel
		nonce := uint32(0) //
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
			case <-i.stopper.ShouldStop():
				return
			default:
			}
		}
	}
}

// Given a particular blockHash, generate a new state by walking through the blockchain
// Automatically adds states if they do not exist already
// TODO: Complete this
func (i *InkMiner) CalculateState(blockHash string) (newState State, err error) {
	newState = State{}
	newState.shapes = make(map[string]blockartlib.Shape)
	newState.shapeOwners = make(map[string]string)
	newState.inkLevels = make(map[string]uint32)

	err = nil

	i.mu.Lock()
	defer i.mu.Unlock()

	// Check if the block exists on the blockchain, fail if not found
	block, ok := i.mu.blockchain[blockHash]
	if !ok {
		i.log.Println("Invalid blockhash")
		return newState, blockartlib.InvalidBlockHashError(blockHash)
	}

	if block.PrevBlock == "" {
		// Invalid block, is it a genesis block
		return newState, nil
	}

	if block.PrevBlock == i.settings.GenesisBlockHash {
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
			i.log.Println("Invalid blockhash")
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

	// Check if we did not find a block to work with, if so generate a "blank state"
	if !foundState {
		lastState = State{}
		lastState.shapes = make(map[string]blockartlib.Shape)
		lastState.shapeOwners = make(map[string]string)
		lastState.inkLevels = make(map[string]uint32)
	}

	// Now, attempt to work through the worklist
	for pos := len(workListStack) - 1; pos > 0; pos-- {
		workingBlock := workListStack[pos]
		createdState := lastState

		// For each operation, add each entry
		for _, op := range workingBlock.Records {
			pubkey := op.PubKey
			opCost := op.InkCost
			shape := op.Shape
			shapeHash := op.ShapeHash
			createdState.shapes[shapeHash] = shape
			createdState.shapeOwners[shapeHash] = pubkey
			if createdState.inkLevels[pubkey] > opCost {
				createdState.inkLevels[pubkey] = createdState.inkLevels[pubkey] - opCost
			} else {
				fmt.Println("Ink levels somehow became lower than 0...")
				createdState.inkLevels[pubkey] = 0
			}
		}
		workingBlockHash, err := workingBlock.Hash()
		if err != nil {
			fmt.Println("Error hashing block")
		}

		// Gives reward based on the block
		if len(workingBlock.Records) > 0 {
			// Operation block
			createdState.inkLevels[workingBlockHash] += i.settings.InkPerOpBlock
		} else {
			// NoOp block
			createdState.inkLevels[workingBlockHash] += i.settings.InkPerNoOpBlock
		}

		// Update the states
		i.states[workingBlockHash] = createdState
		// Set the state walker with the newest state
		lastState = createdState
	}

	newState, ok = i.states[blockHash]
	if !ok {
		i.log.Println("This should never occur...")
		return newState, blockartlib.InvalidBlockHashError(blockHash)
	}

	return newState, err
}

func (i *InkMiner) retrieveState(blockHash string) (retState State, err error) {
	if retState, ok := i.states[blockHash]; !ok {
		return State{}, blockartlib.InvalidBlockHashError(blockHash)
	} else {
		return retState, nil
	}
}

// Starts from the
//func (i *InkMiner) autoAddStates() bool {
//
//}
