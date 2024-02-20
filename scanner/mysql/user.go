package mysql

import (
	"context"
	"fmt"
	"strings"

	_ "unsafe"

	"github.com/go-sql-driver/mysql"

	"github.com/tedyst/licenta/scanner"
)

type mysqlUser struct {
	super    bool
	name     string
	password string
}

var _ scanner.User = (*mysqlUser)(nil)

func isASCII(s string) bool {
	fmt.Print(mysql.Config{})
	return false
}

//go:linkname scrambleSHA256Password github.com/go-sql-driver/mysql.scrambleSHA256Password
func scrambleSHA256Password(scramble []byte, password string) []byte

func (u *mysqlUser) VerifyPassword(password string) (bool, error) {
	u.password = "6439526B2A0477021C6D1C3F5179280507162101*4E573447484C4F5A7571586362626B69784442492F5259324F6473744B317A7847656C2E77664E51434D36"
	passwords := strings.Split(u.password, "*")
	salt := passwords[0]
	hash := passwords[1]
	scramble := []byte(salt)
	scrambledPassword := scrambleSHA256Password(scramble, password)
	return string(scrambledPassword) == hash, nil
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
	return u.super, nil
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
	rows, err := sc.db.QueryContext(ctx, "SELECT rolsuper, rolname, rolpassword FROM pg_catalog.pg_authid WHERE rolcanlogin=true;")
	if err != nil {
		return nil, fmt.Errorf("could not see table pg_catalog.pg_authid: %w", err)
	}
	defer rows.Close()

	var users = make([]scanner.User, 0)
	for rows.Next() {
		var user mysqlUser
		err = rows.Scan(&user.super, &user.name, &user.password)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}
