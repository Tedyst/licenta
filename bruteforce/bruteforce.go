package bruteforce

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/scanner"
	"golang.org/x/sync/semaphore"
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

func BruteforcePasswordAllUsers(ctx context.Context, sc scanner.Scanner, database db.TransactionQuerier, origStatusFunc StatusFunc) ([]scanner.ScanResult, error) {
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
	statusLock := sync.Mutex{}
	for _, user := range users {
		status[user] = BruteforceUserStatus{
			Total: count,
			Tried: 0,
		}
	}

	statusFunc := func(status map[scanner.User]BruteforceUserStatus) error {
		statusLock.Lock()
		defer statusLock.Unlock()
		return origStatusFunc(status)
	}

	err = statusFunc(status)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		pass, ok, err := bruteforcePasswordsUser(ctx, user, database, status, statusFunc, &statusLock)
		if err != nil {
			return nil, err
		}
		err = statusFunc(status)
		if err != nil {
			return nil, err
		}
		username, err := user.GetUsername()
		if err != nil {
			return nil, err
		}
		hashed, err := user.GetHashedPassword()
		if err != nil {
			return nil, err
		}

		u, exists := status[user]
		if !exists {
			return nil, errors.New("user not found in status")
		}

		err = database.InsertBruteforcedPassword(ctx, queries.InsertBruteforcedPasswordParams{
			Username: username,
			Password: sql.NullString{String: pass, Valid: pass != ""},
			Hash:     hashed,
			LastBruteforceID: sql.NullInt64{
				Int64: int64(u.Total),
				Valid: true,
			},
		})
		if err != nil {
			return nil, err
		}

		if ok {
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
	statusLock *sync.Mutex,
) (string, bool, error) {
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
		statusLock.Lock()
		defer statusLock.Unlock()
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

	hash, err := u.GetHashedPassword()
	if err != nil {
		return "", false, err
	}
	username, err := u.GetUsername()
	if err != nil {
		return "", false, err
	}

	var startBruteforceID int64 = 0

	if hash != "" {
		alreadySolved, err := database.GetBruteforcedPasswordByHashAndUsername(ctx, queries.GetBruteforcedPasswordByHashAndUsernameParams{
			Hash:     hash,
			Username: username,
		})
		if err != nil && err != pgx.ErrNoRows {
			return "", false, err
		}
		if alreadySolved != nil && alreadySolved.Password.Valid {
			statusLock.Lock()
			defer statusLock.Unlock()
			if entry, ok := status[u]; ok {
				entry.Tried = entry.Total
				entry.FoundPassword = alreadySolved.Password.String
				status[u] = entry
			}
			return alreadySolved.Password.String, true, nil
		}
		if alreadySolved != nil {
			startBruteforceID = alreadySolved.ID
		}
	}

	rows, err := database.GetRawPool().Query(ctx, "SELECT password FROM default_bruteforce_passwords WHERE id >= $1 ORDER BY id DESC", startBruteforceID)
	if err != nil {
		return "", false, err
	}
	defer rows.Close()

	ticker := time.NewTicker(1 * time.Second)
	errorChan := make(chan error, 1)
	resultChan := make(chan string, 1)

	sm := semaphore.NewWeighted(10)
	wg := sync.WaitGroup{}

	for rows.Next() {
		if rows.Err() != nil {
			return "", false, rows.Err()
		}
		var password string
		err = rows.Scan(&password)
		if err != nil {
			return "", false, err
		}

		sm.Acquire(ctx, 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer sm.Release(1)

			ok, err := u.VerifyPassword(password)
			if err != nil {
				errorChan <- err
			}

			statusLock.Lock()
			defer statusLock.Unlock()

			if ok {
				if entry, ok := status[u]; ok {
					entry.FoundPassword = password
					status[u] = entry
				}
				resultChan <- password
				return
			}

			if entry, ok := status[u]; ok {
				entry.Tried += 1
				status[u] = entry
			}
		}()

		select {
		case <-ticker.C:
			statusFunc(status)
		case err := <-errorChan:
			return "", false, err
		case pass := <-resultChan:
			return pass, true, nil
		default:
		}
	}

	wg.Wait()

	statusLock.Lock()
	defer statusLock.Unlock()

	if entry, ok := status[u]; ok {
		entry.Tried = entry.Total
		status[u] = entry
	}

	select {
	case err := <-errorChan:
		return "", false, err
	case pass := <-resultChan:
		return pass, true, nil
	default:
	}
	statusFunc(status)

	return "", false, err
}
