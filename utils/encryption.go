package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

var KeySize = 3072
var SaltSize = 32                  // in bytes
var NonceSize = 24                 // in bytes. taken from aead.NonceSize()
var EncryptionKeySize = uint32(32) // KeySize is 32 bytes (256 bits).
var KeyTime = uint32(5)
var KeyMemory = uint32(1024 * 64) // KeyMemory in KiB. here, 64 MiB.
var KeyThreads = uint8(4)
var ChunkSize = 1024 * 32 // chunkSize in bytes. here, 32 KiB.

func EncryptWithAES(dek []byte, filePath string, encryptedFilePath string) error {
	// Reading file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	buffer := make([]byte, ChunkSize)
	ad_counter := 0 // associated data is a counter

	// Creating block of algorithm
	block, err := aes.NewCipher(dek)
	if err != nil {
		log.Fatalf("cipher err: %v", err.Error())
		return err
	}

	// Creating GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("cipher GCM err: %v", err.Error())
		return err
	}

	for {
		bytesread, err := file.Read(buffer)

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}

			break
		}

		// Generating random nonce
		nonce := make([]byte, gcm.NonceSize(), gcm.NonceSize()+bytesread+gcm.Overhead())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			log.Fatalf("nonce  err: %v", err.Error())
			return err
		}

		// Decrypt file
		cipherText := gcm.Seal(nonce, nonce, buffer[:bytesread], []byte(string(ad_counter)))

		f, err := os.OpenFile(encryptedFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		if _, err = f.Write(cipherText); err != nil {
			panic(err)
		}

		ad_counter += 1
		fmt.Println("bytes read: ", bytesread)
	}

	return nil
}

func DecryptWithAES(dek []byte, encryptedFilePath string, decryptedFilePath string) error {
	// Reading encrypted file
	file, err := os.Open(encryptedFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Creating block of algorithm
	block, err := aes.NewCipher(dek)
	if err != nil {
		log.Fatalf("cipher err: %v", err.Error())
		return err
	}

	// Creating GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("cipher GCM err: %v", err.Error())
		return err
	}

	buffer := make([]byte, gcm.NonceSize()+ChunkSize+gcm.Overhead())
	ad_counter := 0 // associated data is a counter

	for {
		bytesread, err := file.Read(buffer)

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}

			break
		}

		encryptedMsg := buffer[:bytesread]
		// Decrypt file
		nonce, ciphertext := encryptedMsg[:gcm.NonceSize()], encryptedMsg[gcm.NonceSize():]
		plainText, err := gcm.Open(nil, nonce, ciphertext, []byte(string(ad_counter)))
		if err != nil {
			return err
		}

		f, err := os.OpenFile(decryptedFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}

		defer f.Close()

		if _, err = f.Write(plainText); err != nil {
			return err
		}

		fmt.Println("decrypt bytes read: ", bytesread)
		ad_counter += 1
	}

	return nil
}

func EncryptWithChacha20poly1305(dek []byte, filePath string, encryptedFilePath string) error {
	salt := make([]byte, SaltSize)
	if n, err := rand.Read(salt); err != nil || n != SaltSize {
		log.Println("Error when generating random salt.")
		return err
	}

	outfile, err := os.OpenFile(encryptedFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println("Error when opening/creating output file.")
		return err
	}
	defer outfile.Close()

	outfile.Write(salt)

	key := argon2.IDKey(dek, salt, KeyTime, KeyMemory, KeyThreads, EncryptionKeySize)

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		log.Println("Error when creating cipher.")
		return err
	}

	infile, err := os.Open(filePath)
	if err != nil {
		log.Println("Error when opening input file.")
		return err
	}
	defer infile.Close()

	buf := make([]byte, ChunkSize)
	ad_counter := 0 // associated data is a counter

	for {
		n, err := infile.Read(buf)

		if n > 0 {
			// Select a random nonce, and leave capacity for the ciphertext.
			nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+n+aead.Overhead())
			if m, err := rand.Read(nonce); err != nil || m != aead.NonceSize() {
				log.Println("Error when generating random nonce :", err)
				log.Println("Generated nonce is of following size. m : ", m)
				return err
			}

			msg := buf[:n]
			// Encrypt the message and append the ciphertext to the nonce.
			encryptedMsg := aead.Seal(nonce, nonce, msg, []byte(string(ad_counter)))
			outfile.Write(encryptedMsg)
			ad_counter += 1
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Println("Error when reading input file chunk :", err)
			return err
		}
	}

	return nil
}

func DecryptWithChacha20poly1305(dek []byte, encryptedFilePath string, decryptedFilePath string) error {
	infile, err := os.Open(encryptedFilePath)
	if err != nil {
		log.Println("Error when opening input file.")
		return err
	}
	defer infile.Close()

	salt := make([]byte, SaltSize)
	n, err := infile.Read(salt)
	if n != SaltSize {
		log.Printf("Error. Salt should be %d bytes long. salt n : %d", SaltSize, n)
		return err
	}
	if err == io.EOF {
		log.Println("Encountered EOF error.")
		return err
	}
	if err != nil {
		log.Println("Error encountered :", err)
		return err
	}

	key := argon2.IDKey(dek, salt, KeyTime, KeyMemory, KeyThreads, EncryptionKeySize)
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		log.Println("Error when creating cipher.")
		return err
	}
	decbufsize := aead.NonceSize() + ChunkSize + aead.Overhead()

	outfile, err := os.OpenFile(decryptedFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println("Error when opening output file.")
		return err
	}
	defer outfile.Close()

	buf := make([]byte, decbufsize)
	ad_counter := 0 // associated data is a counter

	for {
		n, err := infile.Read(buf)
		if n > 0 {
			encryptedMsg := buf[:n]
			if len(encryptedMsg) < aead.NonceSize() {
				log.Println("Error. Ciphertext is too short.")
				return err
			}

			// Split nonce and ciphertext.
			nonce, ciphertext := encryptedMsg[:aead.NonceSize()], encryptedMsg[aead.NonceSize():]
			// Decrypt the message and check it wasn't tampered with.
			plaintext, err := aead.Open(nil, nonce, ciphertext, []byte(string(ad_counter)))
			if err != nil {
				log.Println("Error when decrypting ciphertext. May be wrong password or file is damaged.")
				return err
			}

			outfile.Write(plaintext)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error encountered. Read %d bytes: %v", n, err)
			return err
		}

		ad_counter += 1
	}

	return nil
}
