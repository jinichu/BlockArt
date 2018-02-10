package inkminer

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"../blockartlib"
	"../crypto"
	//"fmt"
	//"crypto/ecdsa"
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
	inkMiner := generateTestInkMiner()

	/*
		PrevBlock string          // Hash of the previous block
		BlockNum  int             // Block number
		Records   []Operation     // Set of operation records
		PubKey    ecdsa.PublicKey // Public key of the InkMiner that mined this block
		Nonce     uint32
	*/

	// Generate Block 1
	newBlock := blockartlib.Block{}
	newBlock.BlockNum = 1
	newBlock.PrevBlock = inkMiner.settings.GenesisBlockHash
	pubKeyValue, err := crypto.UnmarshalPublic(inkMiner.publicKey)
	if err != nil {
		t.Fatal("Error unmarshalling public key")
	}
	newBlock.PubKey = *pubKeyValue

	// Determine a random nonce for the newBlock
	newBlock.Nonce = 4
	blockHash1, err := newBlock.Hash()
	if err != nil {
		t.Fatal("Unable to retrieve the hash of the first TestBlock")
	}
	inkMiner.AddBlock(newBlock)

	// Newer block
	newBlock = blockartlib.Block{}
	newBlock.PrevBlock = blockHash1
	newBlock.BlockNum = 2
	pubKeyValue, err = crypto.UnmarshalPublic(inkMiner.publicKey)
	if err != nil {
		t.Fatal("Error unmarshalling public key")
	}
	newBlock.PubKey = *pubKeyValue
	newBlock.Nonce = 15
	blockHash2, err := newBlock.Hash()
	if err != nil {
		t.Fatal("Unable to retrieve the hash of the first TestBlock")
	}

	// Calculate two blocks at once and test
	someState, err := inkMiner.CalculateState(blockHash2)
	if err != nil {
		t.Fatal("Error encountered when calculating state: ", err)
	}

	// Check if the inkLevels are updated
	if someState.inkLevels[inkMiner.publicKey] != inkMiner.settings.InkPerNoOpBlock * 2 {
		t.Fatal("ERROR: Incorrect inkLevels. Got: ", someState.inkLevels[inkMiner.publicKey],
			" Expected: ", inkMiner.settings.InkPerNoOpBlock * 2)
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

	if state1.inkLevels[inkMiner.publicKey] != inkMiner.settings.InkPerNoOpBlock * 1 {
		t.Fatal("ERROR: Incorrect inkLevels. Got: ", someState.inkLevels[inkMiner.publicKey],
			" Expected: ", inkMiner.settings.InkPerNoOpBlock * 1)
	}

	

	// Create a new Block and append
	block3 := blockartlib.Block{}
	block3.PrevBlock = blockHash2

	// Calculate one block and test

	//newBlock.Nonce =
	//

}

func generateTestInkMiner() *InkMiner {
	/*
		addr      string							// IP Address of the InkMiner
		client    *rpc.Client       				// RPC client to connect to the server
		privKey   *ecdsa.PrivateKey 				// Pub/priv key pair of this InkMiner
		publicKey string							// Public key of the Miner (Note: is this needed?)

		latest        []*blockartlib.Block         	// Latest blocks in the blockchain
		settings      server.MinerNetSettings 		// Settings for this BlockArt network instance
		currentHead   blockartlib.Block       		// Block that InkMiner is mining on (current head)
		mineBlockChan chan blockartlib.Block 		// Channel used to distribute blocks
		rs            *rpc.Server 					// RPC Server
		states        map[string]State 				// States of the canvas at a given block
		stopper       *stopper.Stopper
		log           *log.Logger

		mu struct {
			sync.Mutex

			l               net.Listener
			currentWIPBlock blockartlib.Block
			peers           map[string]*peer

			// blockchain is a map between blockhash and the block
			blockchain map[string]blockartlib.Block
			// all operations that haven't been added to the current block chain
			mempool map[string]blockartlib.Operation
		}
	*/

	rNum := rand.Int()%5 + 1
	var basePath string = "../testkeys/test" + strconv.FormatInt(int64(rNum), 10)

	privKey, err := crypto.LoadPrivate(basePath+"-public.key", basePath+"-private.key")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	inkMiner, err := New(privKey)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return inkMiner
}
