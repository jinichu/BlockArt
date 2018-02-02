package inkminer

import (
	"errors"

	"../blockartlib"
)

var ErrUnimplemented = errors.New("unimplemented")

func (i *InkMiner) floodOperation(operation *blockartlib.Operation) error {
	// TODO: Tristan pls complete
	return ErrUnimplemented
}

func (i *InkMiner) announceBlock(block blockartlib.Block) error {
	return ErrUnimplemented
}
