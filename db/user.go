package db

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/argon2"
)

var PasswordPepper []byte

const argon2Memory = 64 * 1024
const argon2Iterations = 3
const argon2Parallelism = 2
const argon2SaltLength = 16
const argon2KeyLength = 32

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func (u *User) SetPassword(password string) error {
	salt, err := generateRandomBytes(argon2SaltLength)
	if err != nil {
		return err
	}

	hash := argon2.IDKey([]byte(password), append(PasswordPepper, salt...), argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argon2Memory, argon2Iterations, argon2Parallelism, b64Salt, b64Hash)
	u.Password = encodedHash
	return nil
}

func (u *User) VerifyPassword(password string) (bool, error) {
	p, salt, hash, err := decodeHash(u.Password)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), append(PasswordPepper, salt...), p.iterations, p.memory, p.parallelism, p.keyLength)

	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeHash(encodedHash string) (p *params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (u *User) VerifyTOTP(code string) bool {
	if !u.TotpSecret.Valid {
		return false
	}
	return totp.Validate(code, u.TotpSecret.String)
}

func (u *User) GenerateTOTPSecret() error {
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "licenta",
		AccountName: u.Username,
	})
	if err != nil {
		return err
	}
	u.TotpSecret.Valid = true
	u.TotpSecret.String = secret.Secret()
	return nil
}
