package inkminer

import (
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
	inkMiner := generateTestInkMiner(t)

	// Generate Block 1
	block1 := blockartlib.Block{
		BlockNum:  1,
		PrevBlock: inkMiner.settings.GenesisBlockHash,
		PubKey:    inkMiner.privKey.PublicKey,
		Nonce:     4,
	}
	blockHash1, err := block1.Hash()
	if err != nil {
		t.Fatal(err)
	}
	inkMiner.AddBlock(block1)

	// Newer block
	operation1 := blockartlib.Operation{
		OpType:  blockartlib.ADD,
		InkCost: 5,
		Id:      10,
		PubKey:  inkMiner.privKey.PublicKey,
	}
	block2 := blockartlib.Block{
		Records:   []blockartlib.Operation{operation1},
		PrevBlock: blockHash1,
		BlockNum:  2,
		PubKey:    inkMiner.privKey.PublicKey,
		Nonce:     15,
	}
	blockHash2, err := block2.Hash()
	if err != nil {
		t.Fatal(err)
	}

	inkMiner.AddBlock(block2)

	// Calculate two blocks at once and test
	someState, err := inkMiner.CalculateState(block2)
	if err != nil {
		t.Fatal("Error encountered when calculating state: ", err)
	}

	// Check if the inkLevels are updated
	want := inkMiner.settings.InkPerNoOpBlock*1 + inkMiner.settings.InkPerOpBlock*1 - operation1.InkCost
	out := someState.inkLevels[inkMiner.publicKey]
	if out != want {
		t.Fatal("ERROR: Incorrect inkLevels. Got: ", out, " Expected: ", want)
	}

	// Check if the inkMiner contains the block
	if _, ok := inkMiner.states[blockHash2]; !ok {
		t.Log("ERROR: InkMiner has not saved the state to it's map")
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
	operation2 := blockartlib.Operation{
		OpType:  blockartlib.ADD,
		InkCost: 5,
		PubKey:  inkMiner.privKey.PublicKey,
	}
	block3 := blockartlib.Block{
		PrevBlock: blockHash2,
		BlockNum:  3,
		Records:   []blockartlib.Operation{operation2},
		PubKey:    inkMiner.privKey.PublicKey,
		Nonce:     22441,
	}

	block3Hash, err := block3.Hash()
	if err != nil {
		t.Fatal("ERROP: Unable to hash Block3")
	}

	success, err := inkMiner.AddBlock(block3)
	if err != nil || success == false {
		t.Fatal("Unable to add block3 to the blockchain")
	}

	// Check if the state was computed correctly

	if _, err := inkMiner.CalculateState(block3); err != nil {
		t.Fatal(err)
	}

	state3, ok := inkMiner.states[block3Hash]
	if !ok {
		t.Fatal("Block State 3 was not stored correctly")
	}

	{
		got := state3.inkLevels[inkMiner.publicKey]
		want := inkMiner.settings.InkPerNoOpBlock*1 + inkMiner.settings.InkPerOpBlock*2 - operation1.InkCost - operation2.InkCost
		if got != want {
			t.Fatal("ERROR: Incorrect inkLevels. Got: ", got, " Expected: ", want)
		}
	}
}

func generateTestInkMiner(t *testing.T) *InkMiner {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	inkMiner, err := New(privKey)
	if err != nil {
		t.Fatal(err)
	}

	inkMiner.settings = server.MinerNetSettings{
		MinerSettings: server.MinerSettings{
			GenesisBlockHash:       "genesis!",
			InkPerOpBlock:          20,
			InkPerNoOpBlock:        5,
			PoWDifficultyNoOpBlock: 1,
			PoWDifficultyOpBlock:   2,
		},
	}

	return inkMiner
}
