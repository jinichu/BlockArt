package inkminer

import (
	"strings"
	"testing"

	"../blockartlib"
	"../crypto"
)

func setup() (inkMinerRPC InkMinerRPC, err error) {
	inkMiner := &InkMiner{
		publicKey: "public",
	}
	inkMiner.mu.states = make(map[string]State)

	block1 := blockartlib.Block{
		PrevBlock: "1234",
		Nonce:     2,
	}
	block1Hash, err := block1.Hash()
	if err != nil {
		return InkMinerRPC{}, err
	}
	state1 := State{
		shapes:      make(map[string]blockartlib.Shape),
		shapeOwners: make(map[string]string),
		inkLevels:   make(map[string]uint32),
	}
	inkMiner.mu.states[block1Hash] = state1
	block2 := blockartlib.Block{
		PrevBlock: block1Hash,
		Nonce:     3,
	}
	block2Hash, err := block2.Hash()
	if err != nil {
		return InkMinerRPC{}, err
	}
	state2 := State{
		shapes:      make(map[string]blockartlib.Shape),
		shapeOwners: make(map[string]string),
		inkLevels:   make(map[string]uint32),
	}
	inkMiner.mu.states[block2Hash] = state2
	block3 := blockartlib.Block{
		PrevBlock: block2Hash,
		Nonce:     3,
	}
	if err != nil {
		return InkMinerRPC{}, err
	}
	state3 := State{
		shapes:      make(map[string]blockartlib.Shape),
		shapeOwners: make(map[string]string),
		inkLevels:   make(map[string]uint32),
	}
	shape := blockartlib.Shape{
		Svg: "M 0 0 H 20 V 20 H -20 Z",
	}
	shapeHash, err := crypto.Hash(shape)
	if err != nil {
		return InkMinerRPC{}, err
	}
	op := blockartlib.Operation{
		OpType: blockartlib.ADD,
	}
	op.ADD.Shape = shape

	op2 := blockartlib.Operation{
		OpType: blockartlib.ADD,
	}
	op2.ADD.Shape = blockartlib.Shape{
		Svg: "qwerasdf",
	}
	block3.Records = append(block3.Records, op)
	block3.Records = append(block3.Records, op2)
	block3Hash, err := block3.Hash()

	state3.shapes[shapeHash] = shape
	state3.inkLevels["public"] = 50

	block4 := blockartlib.Block{
		PrevBlock: block2Hash,
		Nonce:     14,
	}
	block4Hash, err := block4.Hash()
	if err != nil {
		return InkMinerRPC{}, err
	}

	inkMiner.mu.states[block3Hash] = state3
	inkMiner.mu.currentHead = block3
	inkMiner.mu.blockchain = make(map[string]blockartlib.Block)
	inkMiner.mu.blockchain[block1Hash] = block1
	inkMiner.mu.blockchain[block2Hash] = block2
	inkMiner.mu.blockchain[block3Hash] = block3
	inkMiner.mu.blockchain[block4Hash] = block4
	inkMiner.settings.GenesisBlockHash = "1234"
	inkMinerRPC = InkMinerRPC{
		i: inkMiner,
	}

	return inkMinerRPC, nil
}

func TestGetSvgString(t *testing.T) {
	i, err := setup()
	if err != nil {
		t.Error(err)
	}
	args, err := crypto.Hash("M 0 0 H 50 V 20 H -20 Z")
	if err != nil {
		t.Error(err)
	}
	var resp string
	err = i.GetSvgString(&args, &resp)
	if err == nil {
		t.Fatal("This ShapeHash shouldn't exist")
	}
	shape := blockartlib.Shape{
		Svg: "M 0 0 H 20 V 20 H -20 Z",
	}
	args, err = crypto.Hash(shape)
	if err != nil {
		t.Error(err)
	}
	err = i.GetSvgString(&args, &resp)
	if err != nil || !strings.Contains(resp, "M 0 0 H 20 V 20 H -20 Z") {
		t.Fatal(resp)
	}
}

func TestGetInk(t *testing.T) {
	i, err := setup()
	if err != nil {
		t.Error(err)
	}
	var resp uint32
	var args string
	err = i.GetInk(&args, &resp)
	if err != nil || resp != 50 {
		t.Fatal(err)
	}
}

func TestGetShapes(t *testing.T) {
	i, err := setup()
	if err != nil {
		t.Error(err)
	}
	args := "1234"
	var resp blockartlib.GetShapesResponse
	err = i.GetShapes(&args, &resp)
	if err == nil {
		t.Fatal("This block shouldn't exist")
	}
	args, err = i.i.mu.currentHead.Hash()
	if err != nil {
		t.Error(err)
	}
	err = i.GetShapes(&args, &resp)
	if err != nil || len(resp.ShapeHashes) != 2 {
		t.Fatal(err)
	}
}

func TestGetGenesisBlock(t *testing.T) {
	i, err := setup()
	if err != nil {
		t.Error(err)
	}
	var args string
	var resp string
	err = i.GetGenesisBlock(&args, &resp)
	if err != nil || resp != "1234" {
		t.Fatal(err)
	}
}

func TestGetChildrenBlocks(t *testing.T) {
	i, err := setup()
	if err != nil {
		t.Error(err)
	}
	args := "qwer"
	var resp blockartlib.GetChildrenResponse
	err = i.GetChildrenBlocks(&args, &resp)
	if err == nil {
		t.Fatal("This block shouldn't exist")
	}
	args, err = i.i.mu.currentHead.Hash()
	if err != nil {
		t.Error(err)
	}
	err = i.GetChildrenBlocks(&args, &resp)
	if err != nil || len(resp.BlockHashes) != 0 {
		t.Fatal(err)
	}
	// Block with only one branch
	block1 := blockartlib.Block{
		PrevBlock: "1234",
		Nonce:     2,
	}
	block1Hash, err := block1.Hash()
	if err != nil {
		t.Error(err)
	}
	err = i.GetChildrenBlocks(&block1Hash, &resp)
	if err != nil || len(resp.BlockHashes) != 1 {
		t.Fatal(err)
	}
	// Block with two branches
	block2 := blockartlib.Block{
		PrevBlock: block1Hash,
		Nonce:     3,
	}
	block2Hash, err := block2.Hash()
	if err != nil {
		t.Error(err)
	}
	err = i.GetChildrenBlocks(&block2Hash, &resp)
	if err != nil || len(resp.BlockHashes) != 2 {
		t.Fatal(err)
	}
}
