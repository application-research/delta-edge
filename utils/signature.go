package utils

import (
	"encoding/base64"
	"golang.org/x/crypto/nacl/sign"
)

func verifySignature(message string, signature string, publicKey string) bool {
	messageBytes := []byte(message)
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return false
	}

	signedMessageBytes := append(signatureBytes, messageBytes...)
	var publicKeyArray [32]byte
	copy(publicKeyArray[:], publicKeyBytes)
	_, ok := sign.Open(nil, signedMessageBytes, &publicKeyArray)
	return ok
}
