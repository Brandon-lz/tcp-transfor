package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
)

var nonce = make([]byte, 12)
var aesgcm cipher.AEAD

func AESInit() {
	// When decoded the key should be 16 bytes (AES-128) or 32 (AES-256).
	key, _ := hex.DecodeString("6368616e676520746869732070617373776f726420746f206120736563726574")
	// plaintext := []byte("exampleplaintext")

	// fmt.Println("plaintext: ", string(plaintext))

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce = make([]byte, 12)
	// if _, err := io.ReadFull(rand.Reader, nonce); err != nil {         // 产生一个随机的nonce，生产中要固定
	// 	panic(err.Error())
	// }
	nonce = []byte{62, 154, 52, 216, 38, 21, 77, 63, 226, 111, 251, 236}

	aesgcm, err = cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	// ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	// fmt.Printf("%x\n", ciphertext)

	// 	plaintext2, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}

	// 	fmt.Println("plaintext2: ", string(plaintext2))
}

func AESEncrypt(plaintext []byte) []byte {
	src := aesgcm.Seal(nil, nonce, plaintext, nil)
	// buf := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	// base64.StdEncoding.Encode(buf, src)
	
	return []byte(base64.StdEncoding.EncodeToString(src))
}

func AESDecrypt(ciphertextbytes []byte) ([]byte, error) {
	ciphertextdata, err := base64.StdEncoding.DecodeString(string(ciphertextbytes))
	if err != nil{
		return nil,err
	}
	return aesgcm.Open(nil, nonce, ciphertextdata, nil)
}
