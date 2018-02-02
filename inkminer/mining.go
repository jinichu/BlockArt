package inkminer

import "../blockartlib"

func (i *InkMiner) mineBlock(operation *blockartlib.Operation) (block blockartlib.Block, err error) {
	// TODO: Jonathan - verify operation and start mining this block. Return mined Block

	// This maybe should be structured as "daemon" ie. an infinite for loop with
	// channels in/out so it's possible to interrupt mid block. - Tristan

	return blockartlib.Block{}, nil
}
