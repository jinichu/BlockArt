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
		{"sdfasdf0", 1},
		{"sdfasdf00", 2},
		{"000", 3},
		{"000a", 0},
	}

	for i, c := range cases {
		out := numZeros(c.in)
		if out != c.want {
			t.Errorf("%d. numZeros(%s) = %d; wanted %d", i, c.in, out, c.want)
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

	block.Nonce = nonce
	hash, err := block.Hash()
	if err != nil {
		t.Fatal(err)
	}

	if !found {
		t.Fatalf("expected hash to be found! nonce %d, hash %q", nonce, hash)
	}
	if numZeros(hash) != int(i.settings.PoWDifficultyNoOpBlock) {
		t.Fatalf("expected %d zeros", i.settings.PoWDifficultyNoOpBlock)
	}
}

func TestInkMiner_CalculateState(t *testing.T) {
	inkMiner := generateTestInkMiner(t)

	// Generate Block 1
	block1 := inkMiner.TestMine(t, blockartlib.Block{
		BlockNum:  1,
		PrevBlock: inkMiner.settings.GenesisBlockHash,
		PubKey:    inkMiner.privKey.PublicKey,
		Nonce:     4,
	})
	blockHash1, err := block1.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := inkMiner.AddBlock(block1); err != nil {
		t.Fatal(err)
	}

	// Newer block
	operation1 := blockartlib.Operation{
		OpType: blockartlib.ADD,
		Id:     10,
		PubKey: inkMiner.privKey.PublicKey,
	}
	operation1.ADD.Shape = blockartlib.TestShape(5, 0)

	block2 := inkMiner.TestMine(t, blockartlib.Block{
		Records:   []blockartlib.Operation{operation1},
		PrevBlock: blockHash1,
		BlockNum:  2,
		PubKey:    inkMiner.privKey.PublicKey,
		Nonce:     15,
	})
	blockHash2, err := block2.Hash()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := inkMiner.AddBlock(block2); err != nil {
		t.Fatal(err)
	}

	// Calculate two blocks at once and test
	someState, err := inkMiner.CalculateState(block2)
	if err != nil {
		t.Fatal("Error encountered when calculating state: ", err)
	}

	// Check if the inkLevels are updated
	want := inkMiner.settings.InkPerNoOpBlock*1 + inkMiner.settings.InkPerOpBlock*1 - 5
	out := someState.inkLevels[inkMiner.publicKey]
	if out != want {
		t.Fatal("ERROR: Incorrect inkLevels. Got: ", out, " Expected: ", want)
	}

	// Check if the inkMiner contains the block
	if _, ok := inkMiner.mu.states[blockHash2]; !ok {
		t.Log("ERROR: InkMiner has not saved the state to it's map")
	}

	// Check if the first block was computed properly
	state1, ok := inkMiner.mu.states[blockHash1]
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
		OpType: blockartlib.ADD,
		PubKey: inkMiner.privKey.PublicKey,
	}
	operation2.ADD.Shape = blockartlib.TestShape(5, 1)

	block3 := inkMiner.TestMine(t, blockartlib.Block{
		PrevBlock: blockHash2,
		BlockNum:  3,
		Records:   []blockartlib.Operation{operation2},
		PubKey:    inkMiner.privKey.PublicKey,
		Nonce:     22441,
	})

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

	state3, ok := inkMiner.mu.states[block3Hash]
	if !ok {
		t.Fatal("Block State 3 was not stored correctly")
	}

	{
		got := state3.inkLevels[inkMiner.publicKey]
		want := inkMiner.settings.InkPerNoOpBlock*1 + inkMiner.settings.InkPerOpBlock*2 - 10
		if got != want {
			t.Fatal("ERROR: Incorrect inkLevels. Got: ", got, " Expected: ", want)
		}
	}
}

func TestTransformStateIntersectionsMultipleBlocks(t *testing.T) {
	im := generateTestInkMiner(t)

	key2, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	pubKey2, err := crypto.MarshalPublic(&key2.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	operation1 := blockartlib.Operation{
		OpType: blockartlib.ADD,
		Id:     1,
		PubKey: im.privKey.PublicKey,
	}
	operation1.ADD.Shape = blockartlib.TestShape(5, 0)

	operation2 := blockartlib.Operation{
		OpType: blockartlib.ADD,
		Id:     2,
		PubKey: im.privKey.PublicKey,
	}
	operation2.ADD.Shape = blockartlib.TestShape(5, 0)

	operation3 := blockartlib.Operation{
		OpType: blockartlib.ADD,
		Id:     3,
		PubKey: key2.PublicKey,
	}
	operation3.ADD.Shape = blockartlib.TestShape(5, 0)

	state := NewState()
	state.inkLevels[im.publicKey] = 1000000
	state.inkLevels[pubKey2] = 1000000

	block := blockartlib.Block{
		PrevBlock: im.settings.GenesisBlockHash,
		BlockNum:  1,
		Records:   []blockartlib.Operation{operation1},
		PubKey:    im.privKey.PublicKey,
	}
	hash, err := block.Hash()
	if err != nil {
		t.Fatal(err)
	}
	state, err = im.TransformState(state, block)
	if err != nil {
		t.Fatal(err)
	}

	block = blockartlib.Block{
		PrevBlock: hash,
		BlockNum:  2,
		Records:   []blockartlib.Operation{operation2},
		PubKey:    im.privKey.PublicKey,
	}
	hash, err = block.Hash()
	if err != nil {
		t.Fatal(err)
	}
	state, err = im.TransformState(state, block)
	if err != nil {
		t.Fatal(err)
	}

	block = blockartlib.Block{
		PrevBlock: hash,
		BlockNum:  3,
		Records:   []blockartlib.Operation{operation3},
		PubKey:    im.privKey.PublicKey,
	}
	state, err = im.TransformState(state, block)
	if err == nil {
		t.Fatalf("expected error from intersecting operations")
	}
}

func TestTransformStateIntersectionsOneBlock(t *testing.T) {
	im := generateTestInkMiner(t)

	key2, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	pubKey2, err := crypto.MarshalPublic(&key2.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	operation1 := blockartlib.Operation{
		OpType: blockartlib.ADD,
		Id:     1,
		PubKey: im.privKey.PublicKey,
	}
	operation1.ADD.Shape = blockartlib.TestShape(5, 0)

	operation2 := blockartlib.Operation{
		OpType: blockartlib.ADD,
		Id:     1,
		PubKey: im.privKey.PublicKey,
	}
	operation2.ADD.Shape = blockartlib.TestShape(5, 0)

	operation3 := blockartlib.Operation{
		OpType: blockartlib.ADD,
		Id:     2,
		PubKey: key2.PublicKey,
	}
	operation3.ADD.Shape = blockartlib.TestShape(5, 0)

	state := NewState()
	state.inkLevels[im.publicKey] = 1000000
	state.inkLevels[pubKey2] = 1000000

	block := blockartlib.Block{
		PrevBlock: im.settings.GenesisBlockHash,
		BlockNum:  1,
		Records:   []blockartlib.Operation{operation1, operation3},
		PubKey:    im.privKey.PublicKey,
	}
	if _, err := im.TransformState(state, block); err == nil {
		t.Fatalf("expected error from intersecting operations")
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
		GenesisBlockHash:       "genesis!",
		InkPerOpBlock:          20,
		InkPerNoOpBlock:        5,
		PoWDifficultyNoOpBlock: 0,
		PoWDifficultyOpBlock:   0,
	}

	return inkMiner
}
