package mongodb

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/tedyst/licenta/scanner"
	"github.com/xdg-go/scram"
	"go.mongodb.org/mongo-driver/bson"
)

type mongodbUser struct {
	name           string
	password       string
	algorithm      string
	iterationCount int
	storedKey      []byte
	serverKey      []byte
	salt           []byte
}

var _ scanner.User = (*mongodbUser)(nil)

func (u *mongodbUser) VerifyPassword(password string) (bool, error) {
	switch u.algorithm {
	case "SCRAM-SHA-256":
		client, err := scram.SHA256.NewClient(u.name, password, "")
		if err != nil {
			return false, fmt.Errorf("could not create client: %w", err)
		}
		key := client.GetStoredCredentials(scram.KeyFactors{
			Salt:  string(u.salt),
			Iters: u.iterationCount,
		})
		return bytes.Equal(key.StoredKey, u.storedKey) && bytes.Equal(key.ServerKey, u.serverKey), nil
	case "SCRAM-SHA-1":
		pwhash := md5.Sum([]byte(u.name + ":mongo:" + password))
		pwhashHex := fmt.Sprintf("%x", pwhash)
		client, err := scram.SHA1.NewClient(u.name, pwhashHex, "")
		if err != nil {
			return false, fmt.Errorf("could not create client: %w", err)
		}
		key := client.GetStoredCredentials(scram.KeyFactors{
			Salt:  string(u.salt),
			Iters: u.iterationCount,
		})
		return bytes.Equal(key.StoredKey, u.storedKey) && bytes.Equal(key.ServerKey, u.serverKey), nil
	default:
		return false, errors.New("invalid auth plugin")
	}
}

func (u *mongodbUser) GetRawPassword() (string, bool, error) {
	if strings.HasPrefix(u.password, "SCRAM-SHA-256") {
		return "", false, nil
	}
	if strings.HasPrefix(u.password, "md5") {
		return "", false, nil
	}
	return u.password, true, nil
}

func (u *mongodbUser) IsPrivileged() (bool, error) {
	return true, nil
}

func (u *mongodbUser) HasPassword() (bool, error) {
	return true, nil
}

func (u *mongodbUser) GetUsername() (string, error) {
	return u.name, nil
}

func (u *mongodbUser) GetHashedPassword() (string, error) {
	return u.password, nil
}

func (sc *mongodbScanner) GetUsers(ctx context.Context) ([]scanner.User, error) {
	var result bson.M
	err := sc.db.Database("admin").RunCommand(ctx, bson.D{
		{Key: "usersInfo", Value: 1},
		{Key: "showCredentials", Value: true},
	}).Decode(&result)
	if err != nil {
		return nil, err
	}

	users, ok := result["users"].(bson.A)
	if !ok {
		return nil, errors.New("Could not find users")
	}

	var res []scanner.User
	for _, user := range users {
		userMap, ok := user.(bson.M)
		if !ok {
			return nil, errors.New("Could not find user")
		}

		username, ok := userMap["user"].(string)
		if !ok {
			return nil, errors.New("Could not find username")
		}

		credentials, ok := userMap["credentials"].(bson.M)
		if !ok {
			return nil, errors.New("Could not find credentials")
		}

		for t, credential := range credentials {
			credentialMap, ok := credential.(bson.M)
			if !ok {
				return nil, errors.New("Could not find credential")
			}

			iterationCount, ok := credentialMap["iterationCount"].(int32)
			if !ok {
				return nil, errors.New("Could not find iterationCount")
			}

			storedKey, ok := credentialMap["storedKey"].(string)
			if !ok {
				return nil, errors.New("Could not find storedKey")
			}

			decodedStoredKey, err := base64.StdEncoding.DecodeString(storedKey)
			if err != nil {
				return nil, fmt.Errorf("could not decode storedKey: %w", err)
			}

			serverKey, ok := credentialMap["serverKey"].(string)
			if !ok {
				return nil, errors.New("Could not find serverKey")
			}

			decodedServerKey, err := base64.StdEncoding.DecodeString(serverKey)
			if err != nil {
				return nil, fmt.Errorf("could not decode serverKey: %w", err)
			}

			salt, ok := credentialMap["salt"].(string)
			if !ok {
				return nil, errors.New("Could not find salt")
			}

			decodedSalt, err := base64.StdEncoding.DecodeString(salt)
			if err != nil {
				return nil, fmt.Errorf("could not decode salt: %w", err)
			}

			res = append(res, &mongodbUser{
				name:           username,
				algorithm:      t,
				storedKey:      decodedStoredKey,
				serverKey:      decodedServerKey,
				iterationCount: int(iterationCount),
				salt:           decodedSalt,
			})
		}
	}

	return res, nil
}
