package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

// var key = []byte("1234567890123456")
var key, _ = hex.DecodeString("6368616e67652074686973207061773776f726420746f2adsf20736563726574")

func AESEncryptWithKey(plaintext []byte) (string, error) {
	// block, err := aes.NewCipher(key)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// // Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	// nonce := make([]byte, 12)
	// if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
	// 	panic(err.Error())
	// }

	// aesgcm, err := cipher.NewGCM(block)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// return string(aesgcm.Seal(nil, nonce, plaintext, nil)), nil
	return AESEncrypt(plaintext, key)
}

func AESDecryptWithKey(ciphertextbytes string) ([]byte, error) {
	// nonce, _ := hex.DecodeString("64a9433eae7ccceee2fc0eda")

	// block, err := aes.NewCipher(key)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// aesgcm, err := cipher.NewGCM(block)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// ciphertext, _ := hex.DecodeString(ciphertextbytes)
	// return aesgcm.Open(nil, nonce, ciphertext, nil)
	return AESDecrypt(ciphertextbytes, key)
}

// AES加密
func AESEncrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// 转为Base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AES解密
func AESDecrypt(ciphertext string, key []byte) ([]byte, error) {
	ciphertextdata, _ := base64.StdEncoding.DecodeString(ciphertext)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertextdata) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertextdata[:aes.BlockSize]
	ciphertextdata = ciphertextdata[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertextdata, ciphertextdata)

	return ciphertextdata, nil
}
