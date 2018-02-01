package blockartlib

import (
	"crypto/ecdsa"
	"errors"
)

func (i *InkMiner) InitConnection(req *ecdsa.PublicKey, resp *CanvasSettings) error {
	// Confirm that this is the right public key for this InkMiner
	if *req != i.privKey.Public() {
		return errors.New("Public key is incorrect for this InkMiner")
	}
	*resp = i.settings.canvasSettings
	return nil
}
