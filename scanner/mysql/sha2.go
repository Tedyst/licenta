package mysql

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

const (
	DIGEST_LEN        = 32
	MIXCHARS          = 32
	ROUNDS_MIN        = 1000
	ROUNDS_MAX        = 100000
	ROUNDS_DEFAULT    = 5000
	ROUNDS_MULTIPLIER = 1000
)

var encoderString = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func b64From24bit(B2, B1, B0 byte, N int, p *[]byte) {
	w := uint32(B2)<<16 | uint32(B1)<<8 | uint32(B0)
	n := N
	for n > 0 {
		*p = append(*p, encoderString[w&0x3f])
		w >>= 6
		n--
	}
}

func myCryptGenhash(plaintext []byte, switchsalt []byte, numRounds int) []byte {
	rounds := ROUNDS_DEFAULT
	if numRounds > ROUNDS_MIN && numRounds < ROUNDS_MAX {
		rounds = numRounds
	}

	var salt []byte
	salt = switchsalt

	cryptAlgMagic := []byte("$6$")
	if len(salt) >= len(cryptAlgMagic) && bytes.Equal(salt[0:len(cryptAlgMagic)], cryptAlgMagic) {
		salt = salt[len(cryptAlgMagic):]
	}

	hashA := sha256.New()
	hashB := sha256.New()
	hashC := sha256.New()
	hashDP := sha256.New()
	hashDS := sha256.New()

	hashA.Write(plaintext)
	hashA.Write(salt)

	hashB.Write(plaintext)
	hashB.Write(salt)
	hashB.Write(plaintext)

	hashBBytes := hashB.Sum(nil)

	var i int
	for i = len(plaintext); i > MIXCHARS; i -= MIXCHARS {
		hashA.Write(hashBBytes[:MIXCHARS])
	}
	hashA.Write(hashBBytes[:i])

	for i = len(plaintext); i > 0; i >>= 1 {
		if i&1 != 0 {
			hashA.Write(hashBBytes[:MIXCHARS])
		} else {
			hashA.Write(plaintext)
		}
	}

	hashABytes := hashA.Sum(nil)

	for i = 0; i < len(plaintext); i++ {
		hashDP.Write(plaintext)
	}
	hashDPBytes := hashDP.Sum(nil)

	Pbuf := make([]byte, len(plaintext))
	Pp := Pbuf
	for i = len(plaintext); i >= MIXCHARS; i -= MIXCHARS {
		copy(Pp, hashDPBytes)
		Pp = Pp[MIXCHARS:]
	}
	copy(Pp, hashDPBytes[:i])

	for i := 0; i < 16+int(hashABytes[0]); i++ {
		hashDS.Write(salt)
	}
	hashDSBytes := hashDS.Sum(nil)

	Sbuf := make([]byte, len(salt))
	Sp := Sbuf
	for i = len(salt); i >= MIXCHARS; i -= MIXCHARS {
		copy(Sp, hashDSBytes)
		Sp = Sp[MIXCHARS:]
	}
	copy(Sp, hashDSBytes[:i])

	for i := 0; i < rounds; i++ {
		hashC.Reset()

		if i&1 != 0 {
			hashC.Write(Pbuf)
		} else {
			if i == 0 {
				hashC.Write(hashABytes[:MIXCHARS])
			} else {
				hashC.Write(hashDPBytes[:MIXCHARS])
			}
		}

		if i%3 != 0 {
			hashC.Write(Sbuf)
		}

		if i%7 != 0 {
			hashC.Write(Pbuf)
		}

		if i&1 != 0 {
			if i == 0 {
				hashC.Write(hashABytes[:MIXCHARS])
			} else {
				hashC.Write(hashDPBytes[:MIXCHARS])
			}
		} else {
			hashC.Write(Pbuf)
		}

		hashDPBytes = hashC.Sum(nil)
	}

	resultingHash := []byte{}
	b64From24bit(hashDPBytes[0], hashDPBytes[10], hashDPBytes[20], 4, &resultingHash)
	b64From24bit(hashDPBytes[21], hashDPBytes[1], hashDPBytes[11], 4, &resultingHash)
	b64From24bit(hashDPBytes[12], hashDPBytes[22], hashDPBytes[2], 4, &resultingHash)
	b64From24bit(hashDPBytes[3], hashDPBytes[13], hashDPBytes[23], 4, &resultingHash)
	b64From24bit(hashDPBytes[24], hashDPBytes[4], hashDPBytes[14], 4, &resultingHash)
	b64From24bit(hashDPBytes[15], hashDPBytes[25], hashDPBytes[5], 4, &resultingHash)
	b64From24bit(hashDPBytes[6], hashDPBytes[16], hashDPBytes[26], 4, &resultingHash)
	b64From24bit(hashDPBytes[27], hashDPBytes[7], hashDPBytes[17], 4, &resultingHash)
	b64From24bit(hashDPBytes[18], hashDPBytes[28], hashDPBytes[8], 4, &resultingHash)
	b64From24bit(hashDPBytes[9], hashDPBytes[19], hashDPBytes[29], 4, &resultingHash)
	b64From24bit(byte(0), hashDPBytes[31], hashDPBytes[30], 3, &resultingHash)

	return resultingHash
}

func verifySHA2Password(hashedPassword string, password string) (bool, error) {
	parts := bytes.Split([]byte(hashedPassword), []byte("$"))

	rounds := ROUNDS_DEFAULT
	r, err := strconv.Atoi(string(parts[3]))
	if err == nil {
		rounds = r * ROUNDS_MULTIPLIER
	}
	if rounds < ROUNDS_MIN || rounds > ROUNDS_MAX {
		rounds = ROUNDS_DEFAULT
	}

	salt := make([]byte, hex.DecodedLen(len(parts[4])))
	_, err = hex.Decode(salt, parts[4])
	if err != nil {
		return false, fmt.Errorf("cannot decode hex salt: %s", parts[4])
	}

	encodedHash := make([]byte, hex.DecodedLen(len(parts[5])))
	_, err = hex.Decode(encodedHash, parts[5])
	if err != nil {
		return false, fmt.Errorf("cannot decode hex hash: %s", parts[5])
	}

	resultingHash := myCryptGenhash([]byte(password), salt, rounds)

	return bytes.Equal(resultingHash, encodedHash), nil
}
