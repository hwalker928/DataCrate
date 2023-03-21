package functions

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"golang.org/x/crypto/scrypt"
	"io"
	"io/ioutil"
	"os"
)

func EncryptFile(filepath string, password string, outputFile string) {
	// read content from your file
	plaintext, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err.Error())
	}

	// this is the key
	key, err := scrypt.Key([]byte(password), []byte("D@t@Cr@tes!!"), 16384, 8, 1, 32)
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore, it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// create a new file for saving the encrypted data.
	f, err := os.Create(outputFile)
	defer f.Close()
	if err != nil {
		panic(err.Error())
	}
	_, err = io.Copy(f, bytes.NewReader(ciphertext))
	if err != nil {
		panic(err.Error())
	}
}

func DecryptFile(filepath string, password string, outputFile string) {
	// read the encrypted content from the file
	ciphertext, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err.Error())
	}

	// extract the IV from the ciphertext
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// this is the key
	key, err := scrypt.Key([]byte(password), []byte("D@t@Cr@tes!!"), 16384, 8, 1, 32)
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	stream := cipher.NewCFBDecrypter(block, iv)

	// decrypt the ciphertext
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	// create a new file for saving the decrypted data.
	f, err := os.Create(outputFile)
	defer f.Close()

	if err != nil {
		panic(err.Error())
	}
	_, err = io.Copy(f, bytes.NewReader(plaintext))
	if err != nil {
		panic(err.Error())
	}
}

func IsValidZipFile(filePath string) bool {
	// Open the zip file for reading
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return false
	}
	defer r.Close()

	// Iterate over each file in the zip archive and check for errors
	for _, f := range r.File {
		if _, err := f.Open(); err != nil {
			return false
		}
	}

	return true
}
