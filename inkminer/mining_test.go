package inkminer

import (
	"testing"
	"../crypto"
	"math/rand"
	"strconv"
	"../blockartlib"
	//"fmt"
	//"crypto/ecdsa"
)

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
	lastBlockHash, err := newBlock.Hash()
	if err != nil {
		t.Fatal("Unable to retrieve the hash of the first TestBlock")
	}

	// Newer block
	newBlock = blockartlib.Block{}
	newBlock.PrevBlock = lastBlockHash
	newBlock.BlockNum = 2
	pubKeyValue, err = crypto.UnmarshalPublic(inkMiner.publicKey)
	if err != nil {
		t.Fatal("Error unmarshalling public key")
	}

	// Calculate two blocks at once and test


	// Create a new Block and append

	// Calculate one block and test




	//newBlock.Nonce =
	//


}



func generateTestInkMiner() *InkMiner{
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

	rNum := rand.Int() % 5 + 1
	var basePath string = "../testkeys/test" + strconv.FormatInt(int64(rNum), 10)

	privKey, err := crypto.LoadPrivate(basePath + "-public.key", basePath + "private.key")
	if err != nil {
		return nil
	}

	inkMiner, err := New(privKey)
	if err != nil {
		return nil
	}

	return inkMiner
}