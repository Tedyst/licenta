package postgres

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/scanner"
	"github.com/xdg-go/scram"
)

type postgresUser struct {
	super    bool
	name     string
	password string
}

var _ scanner.User = (*postgresUser)(nil)

func (u *postgresUser) verifyPasswordSCRAMSHA256(password string) (bool, error) {
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
		return false, errors.Wrap(err, "could not convert iterations to int")
	}
	salt := iterAndSalt[1]

	serverAndStored := strings.Split(parts[2], ":")
	if len(serverAndStored) != 2 {
		return false, errors.New("invalid server and stored key format")
	}
	storedKey, err := base64.StdEncoding.DecodeString(serverAndStored[0])
	if err != nil {
		return false, errors.Wrap(err, "could not decode server key")
	}
	serverKey, err := base64.StdEncoding.DecodeString(serverAndStored[1])
	if err != nil {
		return false, errors.Wrap(err, "could not decode stored key")
	}

	client, err := scram.SHA256.NewClient(u.name, password, "")
	if err != nil {
		return false, errors.Wrap(err, "could not create client")
	}
	decodedSalt, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return false, errors.Wrap(err, "could not decode salt")
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

func (sc *postgresScanner) GetUsers(ctx context.Context) ([]scanner.User, error) {
	rows, err := sc.db.Query(ctx, "SELECT rolsuper, rolname, rolpassword FROM pg_catalog.pg_authid WHERE rolcanlogin=true;")
	if err != nil {
		return nil, errors.Wrap(err, "could not see table pg_catalog.pg_authid")
	}
	defer rows.Close()

	var users = make([]scanner.User, 0)
	for rows.Next() {
		var user postgresUser
		err = rows.Scan(&user.super, &user.name, &user.password)
		if err != nil {
			return nil, errors.Wrap(err, "could not scan row")
		}
		users = append(users, &user)
	}

	return users, nil
}
