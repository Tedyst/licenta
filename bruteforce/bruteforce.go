package bruteforce

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/scanner"
)

type bruteforceResult struct {
	user     string
	password string
}

func (*bruteforceResult) Severity() scanner.Severity {
	return scanner.SEVERITY_HIGH
}

func (b *bruteforceResult) Detail() string {
	return "Found password for user " + b.user + " using bruteforce. Discovered password: " + b.password
}

var _ scanner.ScanResult = (*bruteforceResult)(nil)

type BruteforceUserStatus struct {
	Total         int
	Tried         int
	FoundPassword string
}

type StatusFunc = func(map[scanner.User]BruteforceUserStatus) error
type innerStatusFunc = *BruteforceUserStatus

func BruteforcePasswordAllUsers(ctx context.Context, sc scanner.Scanner, database db.TransactionQuerier, statusFunc StatusFunc) ([]scanner.ScanResult, error) {
	results := []scanner.ScanResult{}

	users, err := sc.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	count := 0

	row := database.GetRawPool().QueryRow(ctx, "SELECT COUNT(*) FROM default_bruteforce_passwords")
	if err := row.Scan(&count); err != nil {
		return nil, err
	}

	status := map[scanner.User]BruteforceUserStatus{}
	for _, user := range users {
		status[user] = BruteforceUserStatus{
			Total: count,
			Tried: 0,
		}
	}

	err = statusFunc(status)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		pass, ok, err := bruteforcePasswordsUser(ctx, user, database, status, statusFunc)
		if err != nil {
			return nil, err
		}
		err = statusFunc(status)
		if err != nil {
			return nil, err
		}
		if ok {
			username, err := user.GetUsername()
			if err != nil {
				return nil, err
			}
			results = append(results, &bruteforceResult{
				user:     username,
				password: pass,
			})
		}
	}

	return results, nil
}

func bruteforcePasswordsUser(
	ctx context.Context,
	u scanner.User,
	database db.TransactionQuerier,
	status map[scanner.User]BruteforceUserStatus,
	statusFunc StatusFunc,
) (string, bool, error) {
	rows, err := database.GetRawPool().Query(ctx, "SELECT password FROM default_bruteforce_passwords")
	if err != nil {
		return "", false, err
	}
	defer rows.Close()

	pass, hasrpw, err := u.GetRawPassword()
	if err != nil {
		return "", false, err
	}
	if hasrpw {
		q, err := database.GetRawPool().Query(ctx, "SELECT password FROM default_bruteforce_passwords WHERE password = ?", pass)
		if err != nil {
			return "", false, err
		}
		defer q.Close()
		p := ""
		err = q.Scan(p)
		if err != nil && err != pgx.ErrNoRows {
			return "", false, err
		}
		if err == nil {
			if entry, ok := status[u]; ok {
				entry.Tried = entry.Total
				entry.FoundPassword = p
				status[u] = entry
			}
			return p, true, nil
		}
		if entry, ok := status[u]; ok {
			entry.Tried = entry.Total
			status[u] = entry
		}
		return p, false, nil
	}

	ticker := time.NewTicker(1 * time.Second)

	for rows.Next() {
		if rows.Err() != nil {
			return "", false, rows.Err()
		}
		var password string
		err = rows.Scan(&password)
		if err != nil {
			return "", false, err
		}

		ok, err := u.VerifyPassword(password)
		if err != nil {
			return "", false, err
		}
		if ok {
			if entry, ok := status[u]; ok {
				entry.Tried = entry.Total
				entry.FoundPassword = password
				status[u] = entry
			}
			return password, true, err
		}

		if entry, ok := status[u]; ok {
			entry.Tried += 1
			status[u] = entry
		}

		select {
		case <-ticker.C:
			statusFunc(status)
		default:
		}
	}

	return "", false, err
}
