package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/nathanaelle/password"
)

// Hash password using MySQL 8+ method (SHA256)
func scrambleSHA256Password(salt []byte, password string) []byte {
	if len(password) == 0 {
		return nil
	}

	// XOR(SHA256(password), SHA256(SHA256(SHA256(password)), salt))

	crypt := sha256.New()
	crypt.Write([]byte(password))
	message1 := crypt.Sum(nil)

	crypt.Reset()
	crypt.Write(message1)
	message1Hash := crypt.Sum(nil)

	crypt.Reset()
	crypt.Write(message1Hash)
	crypt.Write(salt)
	message2 := crypt.Sum(nil)

	for i := range message1 {
		message1[i] ^= message2[i]
	}

	return message1
}

func main() {
	asd := "6439526B2A0477021C6D1C3F5179280507162101$4E573447484C4F5A7571586362626B69784442492F5259324F6473744B317A7847656C2E77664E51434D36"
	passwords := bytes.Split([]byte(asd), []byte("$"))
	salt := make([]byte, hex.DecodedLen(len(passwords[0])))
	_, err := hex.Decode(salt, passwords[0])
	if err != nil {
		panic(err)
	}
	encodedHash := make([]byte, hex.DecodedLen(len(passwords[1])))
	_, err = hex.Decode(encodedHash, passwords[1])
	if err != nil {
		panic(err)
	}
	encoding := base64.NewEncoding("./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz").WithPadding(base64.NoPadding)
	hash := make([]byte, encoding.DecodedLen(len(encodedHash)))
	_, err = encoding.Decode(hash, encodedHash)
	if err != nil {
		panic(err)
	}

	hashedPassword := password.SHA256.Crypt([]byte("cosica"), salt, map[string]interface{}{})
	hashedPassword = strings.Split(hashedPassword, "$")[3]

	fmt.Printf("%v\n", len(salt))
	fmt.Printf("%v\n", string(salt))
	fmt.Printf("%v\n\n", base64.StdEncoding.EncodeToString(salt))
	fmt.Printf("%v\n", len(hash))
	fmt.Printf("%v\n", string(hash))
	fmt.Printf("%v\n", base64.StdEncoding.EncodeToString(hash))
	fmt.Printf("\n")
	fmt.Printf("%v\n", hashedPassword)
}
