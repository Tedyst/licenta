package redis

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"

	"github.com/tedyst/licenta/scanner"
)

var extractRegex = regexp.MustCompile("user ([a-zA-Z0-9]*) on (.*) #([a-zA-Z0-9]*) ")

type redisUser struct {
	name     string
	password string
}

var _ scanner.User = (*redisUser)(nil)

func (u *redisUser) VerifyPassword(password string) (bool, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(password))
	if err != nil {
		return false, err
	}

	h := fmt.Sprintf("%x", hash.Sum(nil))
	return h == u.password, nil
}

func (u *redisUser) GetRawPassword() (string, bool, error) {
	return "", false, nil
}

func (u *redisUser) IsPrivileged() (bool, error) {
	return true, nil
}

func (u *redisUser) HasPassword() (bool, error) {
	return true, nil
}

func (u *redisUser) GetUsername() (string, error) {
	return u.name, nil
}

func (u *redisUser) GetHashedPassword() (string, error) {
	return u.password, nil
}

func (sc *redisScanner) GetUsers(ctx context.Context) ([]scanner.User, error) {
	u := sc.db.Do(ctx, "ACL", "LIST").Val().([]interface{})

	var users []scanner.User
	for _, user := range u {
		matches := extractRegex.FindStringSubmatch(user.(string))
		if len(matches) != 4 {
			return nil, fmt.Errorf("unexpected user format: %s", user)
		}

		users = append(users, &redisUser{
			name:     matches[1],
			password: matches[3],
		})
	}

	return users, nil
}
