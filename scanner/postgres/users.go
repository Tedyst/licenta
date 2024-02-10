package postgres

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"unicode"

	"errors"

	"github.com/tedyst/licenta/scanner"
	"github.com/xdg-go/scram"
)

type postgresUser struct {
	super    bool
	name     string
	password string
}

var _ scanner.User = (*postgresUser)(nil)

func isASCII(s string) bool {
	for _, c := range s {
		if !unicode.IsGraphic(c) {
			return false
		}
	}

	return true
}

func (u *postgresUser) verifyPasswordSCRAMSHA256(password string) (bool, error) {
	if !isASCII(password) {
		slog.Warn("verifyPasswordSCRAMSHA256: password is not ASCII", "password", password)
		return false, nil
	}

	parts := strings.Split(u.password, "$")
	if len(parts) != 3 {
		return false, errors.New("invalid hash format")
	}

	iterAndSalt := strings.Split(parts[1], ":")
	if len(iterAndSalt) != 2 {
		return false, errors.New("invalid iteration and salt format")
	}
	iterations, err := strconv.Atoi(iterAndSalt[0])
	if err != nil {
		return false, fmt.Errorf("could not convert iterations to int: %w", err)
	}
	salt := iterAndSalt[1]

	serverAndStored := strings.Split(parts[2], ":")
	if len(serverAndStored) != 2 {
		return false, errors.New("invalid server and stored key format")
	}
	storedKey, err := base64.StdEncoding.DecodeString(serverAndStored[0])
	if err != nil {
		return false, fmt.Errorf("could not decode server key: %w", err)
	}
	serverKey, err := base64.StdEncoding.DecodeString(serverAndStored[1])
	if err != nil {
		return false, fmt.Errorf("could not decode stored key: %w", err)
	}

	client, err := scram.SHA256.NewClient(u.name, password, "")
	if err != nil {
		return false, fmt.Errorf("could not create client: %w", err)
	}
	decodedSalt, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return false, fmt.Errorf("could not decode salt: %w", err)
	}
	key := client.GetStoredCredentials(scram.KeyFactors{
		Salt:  string(decodedSalt),
		Iters: iterations,
	})
	return bytes.Equal(key.StoredKey, storedKey) && bytes.Equal(key.ServerKey, serverKey), nil
}

func (u *postgresUser) verifyPasswordMD5(password string) (bool, error) {
	if len(u.password) != 35 {
		return false, errors.New("verifyPasswordMD5: invalid password format")
	}
	hashedPass := u.password[3:]

	hash := md5.Sum([]byte(password + u.name))

	return hex.EncodeToString(hash[:]) == hashedPass, nil
}

func (u *postgresUser) VerifyPassword(password string) (bool, error) {
	if strings.HasPrefix(u.password, "SCRAM-SHA-256") {
		return u.verifyPasswordSCRAMSHA256(password)
	}
	if strings.HasPrefix(u.password, "md5") {
		return u.verifyPasswordMD5(password)
	}

	return password == u.password, nil
}

func (u *postgresUser) GetRawPassword() (string, bool, error) {
	if strings.HasPrefix(u.password, "SCRAM-SHA-256") {
		return "", false, nil
	}
	if strings.HasPrefix(u.password, "md5") {
		return "", false, nil
	}
	return u.password, true, nil
}

func (u *postgresUser) IsPrivileged() (bool, error) {
	return u.super, nil
}

func (u *postgresUser) HasPassword() (bool, error) {
	return true, nil
}

func (u *postgresUser) GetUsername() (string, error) {
	return u.name, nil
}

func (u *postgresUser) GetHashedPassword() (string, error) {
	return u.password, nil
}

func (sc *postgresScanner) GetUsers(ctx context.Context) ([]scanner.User, error) {
	rows, err := sc.db.Query(ctx, "SELECT rolsuper, rolname, rolpassword FROM pg_catalog.pg_authid WHERE rolcanlogin=true;")
	if err != nil {
		return nil, fmt.Errorf("could not see table pg_catalog.pg_authid: %w", err)
	}
	defer rows.Close()

	var users = make([]scanner.User, 0)
	for rows.Next() {
		var user postgresUser
		err = rows.Scan(&user.super, &user.name, &user.password)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}
