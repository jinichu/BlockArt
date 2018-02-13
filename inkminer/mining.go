package inkminer

import (
	"errors"
	"fmt"
	"log"

	"../blockartlib"
	"../crypto"
)

// BlockDepth returns the block depth for the given hash. It also memoizes the
// depths into the provided map for performance with repeated calls.
func (i *InkMiner) BlockDepth(hash string, depths map[string]int) (int, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.blockDepthLocked(hash, depths)
}

// blockDepthLocked returns the block depth. It must be locked before calling!
// See BlockDepth.
func (i *InkMiner) blockDepthLocked(hash string, depths map[string]int) (int, error) {
	if hash == i.settings.GenesisBlockHash {
		return 0, nil
	}

	if depths != nil {
		if depth, ok := depths[hash]; ok {
			return depth, nil
		}
	}

	prev, ok := i.mu.blockchain[hash]
	if !ok {
		return 0, blockartlib.InvalidBlockHashError(hash)
	}
	depth, err := i.blockDepthLocked(prev.PrevBlock, depths)
	if err != nil {
		return 0, err
	}
	depth += 1

	if depths != nil {
		depths[hash] = depth
	}

	return depth, nil
}

func (i *InkMiner) BlockWithLongestChain() (string, int, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	depths := map[string]int{}
	max := i.settings.GenesisBlockHash
	maxDepth := 0
	for hash := range i.mu.blockchain {
		depth, err := i.blockDepthLocked(hash, depths)
		if err != nil {
			return "", 0, err
		}
		if depth > maxDepth {
			max = hash
			maxDepth = depth
		}
	}

	return max, 0, nil
}

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
	return errors.New("unimplemented")
}

func (i *InkMiner) generateNewMiningBlock() error {

	prevBlock, depth, err := i.BlockWithLongestChain()
	if err != nil {
		return err
	}

	block := blockartlib.Block{
		PrevBlock: prevBlock,
		BlockNum:  depth + 1,
		PubKey:    i.privKey.PublicKey,
	}

	i.mineBlockChan <- block

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
// INVARIANT: The blockchain has all blocks from 1..n precomputed
func (i *InkMiner) CalculateState(block blockartlib.Block) (newState State, err error) {
	newState = NewState()

	i.mu.Lock()
	defer i.mu.Unlock()

	blockHash, err := block.Hash()
	if err != nil {
		return State{}, err
	}

	if block.PrevBlock == "" {
		// Invalid block, is it a genesis block
		return State{}, errors.New("invalid block: missing PrevBlock")
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
	for err == nil && block.PrevBlock != i.settings.GenesisBlockHash {
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
			if block.PrevBlock == i.settings.GenesisBlockHash {
				// We are at the end of the blockchain, don't bother continuing
				break
			}
			continue
		}
	}

	// Check if we did not find a block to work with, if so generate a "blank state"
	if !foundState {
		lastState = NewState()
	}

	// Now, attempt to work through the worklist
	for pos := len(workListStack) - 1; pos >= 0; pos-- {
		workingBlock := workListStack[pos]
		createdState, err := i.TransformState(lastState, workingBlock)
		if err != nil {
			return State{}, err
		}

		// Update the states
		workingBlockHash, err := workingBlock.Hash()
		if err != nil {
			fmt.Println("Error hashing block")
			return State{}, err
		}
		i.states[workingBlockHash] = createdState
		// Set the state walker with the newest state
		lastState = createdState
	}

	newState, ok := i.states[blockHash]
	if !ok {
		i.log.Println("This should never occur...")
		return newState, blockartlib.InvalidBlockHashError(blockHash)
	}

	return newState, err
}

func (i *InkMiner) TransformState(prev State, block blockartlib.Block) (State, error) {
	createdState := prev.Copy()

	createdState.blockNum += 1
	if createdState.blockNum != block.BlockNum {
		return State{}, fmt.Errorf("expected block to have BlockNum = %d; got %d", createdState.blockNum, block.BlockNum)
	}

	// For each operation, add each entry
	for _, op := range block.Records {
		opHash, err := op.Hash()
		if err != nil {
			return State{}, err
		}

		if _, ok := createdState.commitedOperations[opHash]; ok {
			return State{}, fmt.Errorf("operation has already been committed! %+v", opHash)
		}
		createdState.commitedOperations[opHash] = struct{}{}

		pubkey := op.PubKey
		// TODO compute InkCost and ShapeHash here
		opCost := op.InkCost
		shape := op.Shape
		shapeHash := op.ShapeHash
		createdState.shapes[shapeHash] = shape
		createdState.shapeOwners[shapeHash] = pubkey
		if createdState.inkLevels[pubkey] >= opCost {
			createdState.inkLevels[pubkey] = createdState.inkLevels[pubkey] - opCost
		} else {
			fmt.Println("Ink levels somehow became lower than 0...")
			createdState.inkLevels[pubkey] = 0
			return State{}, fmt.Errorf("%s: ink levels below 0!", pubkey)
		}
	}

	// Gives reward based on the block
	rewardPubKey, err := crypto.MarshalPublic(&block.PubKey)
	if err != nil {
		return State{}, err
	}
	if len(block.Records) > 0 {
		// Operation block
		createdState.inkLevels[rewardPubKey] += i.settings.InkPerOpBlock
	} else {
		// NoOp block
		createdState.inkLevels[rewardPubKey] += i.settings.InkPerNoOpBlock
	}

	return createdState, nil
}

func (i *InkMiner) retrieveState(blockHash string) (retState State, err error) {
	if retState, ok := i.states[blockHash]; !ok {
		return State{}, blockartlib.InvalidBlockHashError(blockHash)
	} else {
		return retState, nil
	}
}
