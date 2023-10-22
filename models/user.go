package models

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/pquerna/otp/totp"
	"github.com/tedyst/licenta/db/queries"
	"golang.org/x/crypto/argon2"
)

var tracer = otel.Tracer("github.com/tedyst/licenta/models")

type User = queries.User
type TotpSecretToken = queries.TotpSecretToken

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

func GenerateHash(ctx context.Context, password string) (string, error) {
	_, span := tracer.Start(ctx, "GenerateHash")
	defer span.End()

	salt, err := generateRandomBytes(argon2SaltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), append(PasswordPepper, salt...), argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argon2Memory, argon2Iterations, argon2Parallelism, b64Salt, b64Hash)
	return encodedHash, nil
}

func SetPassword(ctx context.Context, u *User, password string) error {
	p, err := GenerateHash(ctx, password)
	if err != nil {
		u.Password = p
	}
	return err
}

func VerifyPassword(ctx context.Context, u *queries.User, password string) (bool, error) {
	_, span := tracer.Start(ctx, "VerifyPassword")
	defer span.End()

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

func VerifyTOTP(ctx context.Context, totpSecret *TotpSecretToken, code string) bool {
	_, span := tracer.Start(ctx, "VerifyTOTP")
	defer span.End()

	if !totpSecret.Valid {
		return true
	}
	return totp.Validate(code, totpSecret.TotpSecret)
}

func GenerateTOTPSecret(ctx context.Context) (string, error) {
	_, span := tracer.Start(ctx, "GenerateTOTP")
	defer span.End()

	key, err := totp.Generate(totp.GenerateOpts{})
	if err != nil {
		return "", errors.Wrap(err, "error generating totp key")
	}
	return key.Secret(), nil
}
