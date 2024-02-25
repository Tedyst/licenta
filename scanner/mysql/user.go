package mysql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	_ "unsafe"

	"github.com/tedyst/licenta/scanner"
)

type mysqlUser struct {
	name        string
	password    string
	auth_plugin string
}

var _ scanner.User = (*mysqlUser)(nil)

func (u *mysqlUser) VerifyPassword(password string) (bool, error) {
	switch u.auth_plugin {
	case "caching_sha2_password":
		return verifySHA2Password(u.password, password), nil
	default:
		return false, errors.New("invalid auth plugin")
	}
}

func (u *mysqlUser) GetRawPassword() (string, bool, error) {
	if strings.HasPrefix(u.password, "SCRAM-SHA-256") {
		return "", false, nil
	}
	if strings.HasPrefix(u.password, "md5") {
		return "", false, nil
	}
	return u.password, true, nil
}

func (u *mysqlUser) IsPrivileged() (bool, error) {
	return true, nil
}

func (u *mysqlUser) HasPassword() (bool, error) {
	return true, nil
}

func (u *mysqlUser) GetUsername() (string, error) {
	return u.name, nil
}

func (u *mysqlUser) GetHashedPassword() (string, error) {
	return u.password, nil
}

func (sc *mysqlScanner) GetUsers(ctx context.Context) ([]scanner.User, error) {
	rows, err := sc.db.QueryContext(ctx, "SELECT CONCAT(host, ':', user), plugin, CONCAT('$mysql',LEFT(authentication_string,6),'$',INSERT(HEX(SUBSTR(authentication_string,8)),41,0,'$')) AS hash FROM mysql.user WHERE plugin = 'caching_sha2_password' AND authentication_string NOT LIKE '%INVALIDSALTANDPASSWORD%';")
	if err != nil {
		return nil, fmt.Errorf("could not see table mysql.user: %w", err)
	}
	defer rows.Close()

	var users = make([]scanner.User, 0)
	for rows.Next() {
		var user mysqlUser
		err = rows.Scan(&user.name, &user.auth_plugin, &user.password)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}
