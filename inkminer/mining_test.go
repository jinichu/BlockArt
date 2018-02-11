package inkminer

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"../blockartlib"
	"../crypto"
	"../server"
)

func TestNumZeros(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"aasdfasdf", 0},
		{"0sdfasdf", 1},
		{"00sdfasdf", 2},
		{"000", 3},
		{"a000", 0},
	}

	for i, c := range cases {
		out := numZeros(c.in)
		if out != c.want {
			t.Errorf("%d. numZeros(%s) = %d; wanted %s", i, c.in, out, c.want)
		}
	}
}

func TestMineWorker(t *testing.T) {
	i := &InkMiner{}
	i.settings.PoWDifficultyNoOpBlock = 1
	i.settings.PoWDifficultyOpBlock = 2
	block := blockartlib.Block{}
	nonce, found, err := i.mineWorker(block, 0, 10000)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("expected hash to be found!")
	}
	block.Nonce = nonce
	hash, err := block.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if numZeros(hash) != int(i.settings.PoWDifficultyNoOpBlock) {
		t.Fatalf("expected %d zeros", i.settings.PoWDifficultyNoOpBlock)
	}
}

func TestInkMiner_CalculateState(t *testing.T) {
	inkMiner, err := generateTestInkMiner()
	if err != nil {
		t.Fatal("Unable to create inkMiner due to: ", err)
	}

	// Generate Block 1
	block1 := blockartlib.Block{}
	block1.BlockNum = 1
	block1.PrevBlock = inkMiner.settings.GenesisBlockHash
	pubKeyValue, err := crypto.UnmarshalPublic(inkMiner.publicKey)
	if err != nil {
		t.Fatal("Error unmarshalling public key")
	}
	block1.PubKey = *pubKeyValue

	// Determine a random nonce for the newBlock
	block1.Nonce = 4
	blockHash1, err := block1.Hash()
	if err != nil {
		t.Fatal("Unable to retrieve the hash of the first TestBlock")
	}
	inkMiner.AddBlock(block1)

	// Newer block
	block2 := blockartlib.Block{}
	block2.PrevBlock = blockHash1
	block2.BlockNum = 2
	block2.PubKey = *pubKeyValue
	block2.Nonce = 15
	blockHash2, err := block2.Hash()
	if err != nil {
		t.Fatal("Unable to retrieve the hash of the first TestBlock")
	}

	inkMiner.AddBlock(block2)

	// Calculate two blocks at once and test
	someState, err := inkMiner.CalculateState(blockHash2)
	if err != nil {
		t.Fatal("Error encountered when calculating state: ", err)
	}

	// Check if the inkLevels are updated
	if someState.inkLevels[inkMiner.publicKey] != inkMiner.settings.InkPerNoOpBlock*2 {
		t.Fatal("ERROR: Incorrect inkLevels. Got: ", someState.inkLevels[inkMiner.publicKey],
			" Expected: ", inkMiner.settings.InkPerNoOpBlock*2)
	}

	// Check if the inkMiner contains the block
	if _, ok := inkMiner.states[blockHash2]; !ok {
		t.Log("ERROR: InkMIner has not saved the state to it's map")
	}

	// Check if the first block was computed properly
	state1, ok := inkMiner.states[blockHash1]
	if !ok {
		t.Fatal("Block hash was not computed properly, invariant violated")
	}

	if state1.inkLevels[inkMiner.publicKey] != inkMiner.settings.InkPerNoOpBlock*1 {
		t.Fatal("ERROR: Incorrect inkLevels. Got: ", state1.inkLevels[inkMiner.publicKey],
			" Expected: ", inkMiner.settings.InkPerNoOpBlock*1)
	}

	// Create a new Block and append
	// New block contains a record that has 5 cost
	block3 := blockartlib.Block{}
	block3.PrevBlock = blockHash2
	block3.BlockNum = 3
	operation := blockartlib.Operation{}
	operation.InkCost = 5
	operation.PubKey = inkMiner.publicKey
	block3.Records = []blockartlib.Operation{operation}
	block3.PubKey = *pubKeyValue
	block3.Nonce = 22441

	block3Hash, err := block3.Hash()
	if err != nil {
		t.Fatal("ERROP: Unable to hash Block3")
	}

	success, err := inkMiner.AddBlock(block3)
	if err != nil || success == false {
		t.Fatal("Unable to add block3 to the blockchain")
	}

	// Check if the state was computed correctly

	inkMiner.CalculateState(block3Hash)
	state3, ok := inkMiner.states[block3Hash]
	if !ok {
		t.Fatal("Block State 3 was not stored correctly")
	}

	expectedInkLevels := inkMiner.settings.InkPerNoOpBlock*2 + inkMiner.settings.InkPerOpBlock - 5

	if state3.inkLevels[inkMiner.publicKey] != expectedInkLevels {
		t.Fatal("ERROR: Incorrect inkLevels. Got: ", state3.inkLevels[inkMiner.publicKey],
			" Expected: ", expectedInkLevels)
	}
}

func generateTestInkMiner() (*InkMiner, error) {
	rNum := rand.Int()%5 + 1
	var basePath string = "../testkeys/test" + strconv.FormatInt(int64(rNum), 10)

	// Attempt to load the privatekey from file
	privKey, err := crypto.LoadPrivate(basePath+"-public.key", basePath+"-private.key")
	if err != nil {
		// If we fail the first time, try again with local path
		err = nil
		var basePath string = "./testkeys/test" + strconv.FormatInt(int64(rNum), 10)

		privKey, err := crypto.LoadPrivate(basePath+"-public.key", basePath+"-private.key")
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		inkMiner, err := New(privKey)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		inkMiner.settings = server.MinerNetSettings{}
		inkMiner.settings.GenesisBlockHash = "genesis!"
		inkMiner.settings.InkPerOpBlock = 5
		inkMiner.settings.InkPerNoOpBlock = 10
		inkMiner.settings.PoWDifficultyNoOpBlock = 1
		inkMiner.settings.PoWDifficultyOpBlock = 2
		return inkMiner, nil
	} else {
		inkMiner, err := New(privKey)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		inkMiner.settings = server.MinerNetSettings{}
		inkMiner.settings.GenesisBlockHash = "genesis!"
		inkMiner.settings.InkPerOpBlock = 5
		inkMiner.settings.InkPerNoOpBlock = 10
		inkMiner.settings.PoWDifficultyNoOpBlock = 1
		inkMiner.settings.PoWDifficultyOpBlock = 2

		return inkMiner, nil
	}
}
