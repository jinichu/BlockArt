package inkminer

import (
	"testing"
	"../crypto"
	"math/rand"
	"strconv"
	"../blockartlib"
)

func TestInkMiner_CalculateState(t *testing.T) {

}



func generateTestInkMiner() *InkMiner{
	/*	addr      string							// IP Address of the InkMiner
	client    *rpc.Client       				// RPC client to connect to the server
	privKey   *ecdsa.PrivateKey 				// Pub/priv key pair of this InkMiner
	publicKey string							// Public key of the Miner (Note: is this needed?)

	latest        []*blockartlib.Block         	// Latest blocks in the blockchain
	settings      blockartlib.MinerNetSettings 	// Settings for this BlockArt network instance
	currentHead   *blockartlib.Block           	// Block that InkMiner is mining on (current head)
	mineBlockChan chan blockartlib.Block 		// Channel used to distribute blocks
	rs            *rpc.Server 					// RPC Server
	states        map[string]State 				// States of the canvas at a given block
	stopper       *stopper.Stopper				//
	log           *log.Logger					//

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

	//inkMiner := New()

	inkMiner := new(InkMiner)
	inkMiner.client = nil
	rNum := rand.Int() % 5 + 1
	var basePath string = "../testkeys/test" + strconv.FormatInt(int64(rNum), 10)

	privKey, err := crypto.LoadPrivate(basePath + "-public.key", basePath + "private.key")
	if err != nil {
		return nil
	}
	inkMiner.privKey = privKey
	inkMiner.publicKey, err = crypto.MarshalPublic(&privKey.PublicKey)
	if err != nil {
		return nil
	}

	inkMiner.settings = blockartlib.MinerNetSettings{}
	// Generate genesis block
	genesisBlock := blockartlib.Block{}
	genesisBlock.Nonce = 100 // Default
	genesisBlock.BlockNum = 0
	genesisBlock.Records = []blockartlib.Operation{}
	genesisBlock.PrevBlock = ""

	inkMiner.settings.GenesisBlockHash, err = crypto.Hash(genesisBlock)
	if err != nil {
		return nil
	}
	// TODO: Include other elements for the settings struct?



	return inkMiner
}