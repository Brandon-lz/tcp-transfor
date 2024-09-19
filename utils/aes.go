package utils

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "io"
)

var key = []byte("1234567890123456")

func AESEncryptWithKey(plaintext []byte) (string, error) {
    return AESEncrypt(plaintext, key)
}


func AESDecryptWithKey(ciphertext string) ([]byte, error) {
    return AESDecrypt(ciphertext, key)
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