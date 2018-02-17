package inkminer

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"../blockartlib"
	"../crypto"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.DurationVar(&TestBlockDelay, "delay", 1*time.Second, "mining block delay")
}

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
	var maxBlock blockartlib.Block
	maxDepth := 0

	for hash, block := range i.mu.blockchain {
		depth, err := i.blockDepthLocked(hash, depths)
		if err != nil {
			i.log.Printf("invalid block: %+v: %+v", block, err)
			continue
		}
		if depth > maxDepth {
			max = hash
			maxDepth = depth
			maxBlock = block
		}
	}

	i.mu.currentHead = maxBlock

	return max, maxDepth, nil
}

func (i *InkMiner) GetBlock(hash string) (blockartlib.Block, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()

	b, ok := i.mu.blockchain[hash]
	return b, ok
}

func (i *InkMiner) getStateForHash(hash string) (State, error) {
	if hash == i.settings.GenesisBlockHash {
		return NewState(), nil
	}
	block, ok := i.GetBlock(hash)
	if !ok {
		return State{}, fmt.Errorf("hash not found: %+v", hash)
	}
	state, err := i.CalculateState(block)
	if err != nil {
		return State{}, err
	}
	return state, nil
}

type opError struct {
	blockNum int
	err      error
}

func (i *InkMiner) generateNewMiningBlock() (blockartlib.Block, error) {

	prevBlockHash, _, err := i.BlockWithLongestChain()
	if err != nil {
		return blockartlib.Block{}, err
	}

	state, err := i.getStateForHash(prevBlockHash)
	if err != nil {
		return blockartlib.Block{}, err
	}

	block := blockartlib.Block{
		PrevBlock: prevBlockHash,
		BlockNum:  state.blockNum + 1,
		PubKey:    i.privKey.PublicKey,
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	for hash, op := range i.mu.mempool {
		if _, ok := state.commitedOperations[hash]; ok {
			continue
		}

		// check to see if there have been ValidateNum blocks that have been unable
		// to include the operation. If there have been, send an error to the waiter
		// if it exists.
		firstError, ok := i.mu.opErrors[hash]
		if ok && (firstError.blockNum+int(op.ValidateNum) < block.BlockNum) {
			waiter, ok := i.mu.validateNumMap[hash]
			if ok {
				delete(i.mu.validateNumMap, hash)
				waiter.err <- firstError.err
			}
			continue
		}

		block.Records = append(block.Records, op)
		if _, err := i.TransformState(state, block); err != nil {
			i.log.Printf("op can't be applied to block: %+v, %+v", op, err)
			block.Records = block.Records[:len(block.Records)-1]
			if !ok {
				i.mu.opErrors[hash] = opError{
					blockNum: block.BlockNum,
					err:      err,
				}
			}
			continue
		}
	}

	return block, nil
}

var TestBlockDelay time.Duration

func (i *InkMiner) generateNewMiningBlockLoop(mineBlockChan chan blockartlib.Block) {
	for {
		start := time.Now()
		i.log.Printf("block generate loop")

		// For testing purposes to limit computational cost of block mining.
		if TestBlockDelay > 0 {
			time.Sleep(TestBlockDelay)
		}

		// Clear the newOp/newBlock channels to avoid duplicate work.
	clearLoop:
		for {
			select {
			case <-i.newOpChan:
			case <-i.newBlockChan:
			case <-i.stopper.ShouldStop():
				return
			default:
				break clearLoop
			}
		}

		block, err := i.generateNewMiningBlock()
		if err != nil {
			i.log.Printf("failed to generate new mining block: %+v", err)
			continue
		}
		i.log.Printf("generated block, took %s", time.Since(start))
		mineBlockChan <- block

		// wait for a new operation or block to come in
		select {
		case <-i.newOpChan:
		case <-i.newBlockChan:
		case <-i.stopper.ShouldStop():
			return
		}
	}
}

// startMining should only ever be called once.
func (i *InkMiner) startMining() error {
	mineBlockChan := make(chan blockartlib.Block, 1)

	go i.generateNewMiningBlockLoop(mineBlockChan)
	go i.minerLoop(mineBlockChan)

	return nil
}

// mineBlock returns the nonce, whether or not it found a valid nonce and an
// error.
func (i *InkMiner) mineWorker(block blockartlib.Block, oldNonce uint32, maxIterations int) (uint32, bool, error) {
	difficulty := i.settings.PoWDifficultyOpBlock
	if len(block.Records) == 0 {
		difficulty = i.settings.PoWDifficultyNoOpBlock
	}

	hashNoNonce, err := block.HashNoNonce()
	if err != nil {
		return 0, false, err
	}

	for i := 0; i < maxIterations; i++ {
		oldNonce += 1
		block.Nonce = oldNonce
		hash, err := block.HashApplyNonce(hashNoNonce)
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
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] != '0' {
			return len(str) - i - 1
		}
	}
	return len(str)
}

func (i *InkMiner) minerLoop(blocks <-chan blockartlib.Block) {
	block := <-blocks // Grab a block from the channel
outer:
	for {
		i.log.Printf("mining block...")
		start := time.Now()

		nonce := uint32(rand.Int31())
		found := false
		var err error
		for !found {
			// attempt to mine a block for a set number of iterations
			nonce, found, err = i.mineWorker(block, nonce, 1000)
			if err != nil {
				i.log.Printf("Mining error: %+v", err)
				continue outer
			}

			// Check if there's a new block to mine on instead.
			select {
			case block = <-blocks:
				continue outer
			case <-i.stopper.ShouldStop():
				return
			default:
			}
		}

		block.Nonce = nonce
		i.log.Printf("AddBlock...")
		if _, err := i.AddBlock(block); err != nil {
			i.log.Printf("Mining error: %+v", err)
		} else {
			i.log.Printf("block mined: %+v", block)
		}

		i.log.Printf("block mined. took %s", time.Since(start))

		block = <-blocks
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

	if block.PrevBlock == "" && block.PrevBlock != i.settings.GenesisBlockHash {
		// Invalid block, is it a genesis block
		return State{}, errors.New("invalid block: missing PrevBlock")
	}

	// If the state was already previously calculated, simply return it
	if state, ok := i.mu.states[blockHash]; ok {
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
		_, ok = i.mu.states[nextHash]
		if ok {
			// We did it! We found a block that has some kind of state attached to it!
			// Save this state, we will use it to compute the rest of our worklist
			foundState = true
			lastState = i.mu.states[nextHash]
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
		i.mu.states[workingBlockHash] = createdState
		// Set the state walker with the newest state
		lastState = createdState
	}

	newState, ok := i.mu.states[blockHash]
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
		return State{}, fmt.Errorf("expected block to have BlockNum = %d; got %d\nblock: %+v", createdState.blockNum, block.BlockNum, block)
	}

	// increment the committed time by one
	for k, v := range createdState.commitedOperations {
		createdState.commitedOperations[k] = v + 1
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
		createdState.commitedOperations[opHash] = 0

		pubkey, err := op.PubKeyString()
		if err != nil {
			return State{}, err
		}

		switch op.OpType {
		case blockartlib.ADD:
			opCost, err := op.ADD.Shape.InkCost()
			if err != nil {
				return State{}, err
			}

			inkLevel := createdState.inkLevels[pubkey]
			if inkLevel < opCost {
				return State{}, blockartlib.InsufficientInkError(inkLevel)
			}
			createdState.inkLevels[pubkey] -= opCost

			for shapeHash, owner := range createdState.shapeOwners {
				if owner == pubkey {
					continue
				}

				shape := createdState.shapes[shapeHash]
				if blockartlib.DoesShapeOverlap(shape, op.ADD.Shape) {
					return State{}, blockartlib.ShapeOverlapError(shapeHash)
				}
			}

			createdState.shapes[opHash] = op.ADD.Shape
			createdState.shapeOwners[opHash] = pubkey

		case blockartlib.DELETE:
			shapeHash := op.DELETE.ShapeHash
			owner, ok := createdState.shapeOwners[shapeHash]
			if !ok {
				return State{}, fmt.Errorf("shape doesn't exist")
			}
			shape := createdState.shapes[shapeHash]
			if owner != pubkey {
				return State{}, fmt.Errorf("owner != user: %q != %q", owner, pubkey)
			}
			delete(createdState.shapeOwners, shapeHash)
			delete(createdState.shapes, shapeHash)

			// make deleted shape white
			if shape.Fill != "transparent" {
				shape.Fill = "white"
			}
			if shape.Stroke != "transparent" {
				shape.Stroke = "white"
			}
			createdState.shapes[opHash] = shape

			opCost, err := shape.InkCost()
			if err != nil {
				return State{}, err
			}
			createdState.inkLevels[pubkey] += opCost

		default:
			return State{}, fmt.Errorf("invalid OpType: %+v", op)
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

// TestMine mines a block to completion. Should only be used for testing
// purposes.
func (i *InkMiner) TestMine(t *testing.T, block blockartlib.Block) blockartlib.Block {
	for {
		nonce, success, err := i.mineWorker(block, 0, 10000000000)
		if err != nil {
			t.Fatal(err)
		}
		if success {
			block.Nonce = nonce
			return block
		}
	}
}
