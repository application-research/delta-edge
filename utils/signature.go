package utils

import (
	"crypto/ecdsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
)

const PUBLIC_KEY_PEM = `-----BEGIN PUBLIC KEY-----
MIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQA9TALHFpCEr6Kyac2xL3iyzoQxaZ5
gYiMvl5ChMkV33dKhuuQlomPvfMwguhu2v1qfeK7Y+LwtCEtzAVoHJqisxgAnNTx
1t/aZhmnNgpDdRjiNA5TOfhdistuv86hGZeKUhaeDmjyBKiOrS7xsmmGPGTi++3s
RmmFzt+W9QJEfjehqf0=
-----END PUBLIC KEY-----`

//func VerifySignature(message string, signature string, publicKey string) bool {
//	messageBytes := []byte(message)
//	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
//	if err != nil {
//		return false
//	}
//	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
//	if err != nil {
//		return false
//	}
//
//	signedMessageBytes := append(signatureBytes, messageBytes...)
//	var publicKeyArray [32]byte
//	copy(publicKeyArray[:], publicKeyBytes)
//	_, ok := sign.Open(nil, signedMessageBytes, &publicKeyArray)
//	return ok
//}

func parseEcdsaSignatureDER(signatureDER []byte) (*big.Int, *big.Int, error) {
	var ecdsaSignature struct {
		R, S *big.Int
	}
	if _, err := asn1.Unmarshal(signatureDER, &ecdsaSignature); err != nil {
		return nil, nil, fmt.Errorf("error parsing ECDSA signature: %v", err)
	}
	return ecdsaSignature.R, ecdsaSignature.S, nil
}

func VerifyEcdsaSha512Signature(message interface{}, signatureBase64 string, publicKeyPEM string) (bool, error) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return false, fmt.Errorf("error marshaling JSON: %v", err)
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false, fmt.Errorf("error decoding base64 signature: %v", err)
	}

	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return false, fmt.Errorf("error decoding PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("error parsing public key: %v", err)
	}

	ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("public key is not an ECDSA public key")
	}

	byteSize := (ecdsaPublicKey.Curve.Params().BitSize + 7) / 8
	if len(signatureBytes) > 2*byteSize {
		// Strip ASN.1 encoding if present
		var sig struct {
			R, S *big.Int
		}
		_, err := asn1.Unmarshal(signatureBytes, &sig)
		if err != nil {
			return false, fmt.Errorf("error unmarshaling ASN.1 signature: %v", err)
		}

		hash := sha512.Sum512(messageBytes)

		return ecdsa.Verify(ecdsaPublicKey, hash[:], sig.R, sig.S), nil
	} else if len(signatureBytes) == byteSize {
		// Raw r and s components
		r := new(big.Int).SetBytes(signatureBytes)
		s := new(big.Int).SetBytes(signatureBytes[byteSize:])

		hash := sha512.Sum512(messageBytes)

		return ecdsa.Verify(ecdsaPublicKey, hash[:], r, s), nil
	} else {
		return false, fmt.Errorf("invalid signature length: %d bytes, expected %d or %d bytes", len(signatureBytes), byteSize, 2*byteSize)
	}
}
