package bruteforce

import (
	"context"
	"errors"
	"sync"
	"time"

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
	Total             int
	Tried             int
	FoundPassword     string
	MaximumInternalID int64
}

type StatusFunc = func(map[scanner.User]BruteforceUserStatus) error

type bruteforcer struct {
	passwordProvider PasswordProvider
	scanner          scanner.Scanner
	updateStatus     func() error

	status     map[scanner.User]BruteforceUserStatus
	statusLock sync.Mutex

	users []scanner.User

	results []scanner.ScanResult
}

func NewBruteforcer(passwordProvider PasswordProvider, sc scanner.Scanner, statusFunc StatusFunc) *bruteforcer {
	br := &bruteforcer{
		passwordProvider: passwordProvider,
		scanner:          sc,
		status:           map[scanner.User]BruteforceUserStatus{},
		statusLock:       sync.Mutex{},
	}
	br.updateStatus = func() error {
		br.statusLock.Lock()
		defer br.statusLock.Unlock()
		return statusFunc(br.status)
	}
	return br
}

func (br *bruteforcer) initialize(ctx context.Context) error {
	count, err := br.passwordProvider.GetCount()
	if err != nil {
		return err
	}

	users, err := br.scanner.GetUsers(ctx)
	if err != nil {
		return err
	}
	br.users = users

	br.statusLock.Lock()
	for _, user := range users {
		br.status[user] = BruteforceUserStatus{
			Total: count,
			Tried: 0,
		}
	}
	br.statusLock.Unlock()

	return br.updateStatus()
}

func (br *bruteforcer) savePasswordHash(ctx context.Context, user scanner.User, password string) error {
	username, err := user.GetUsername()
	if err != nil {
		return err
	}
	hash, err := user.GetHashedPassword()
	if err != nil {
		return err
	}
	return br.passwordProvider.SavePasswordHash(username, hash, password, br.status[user].MaximumInternalID)
}

func (br *bruteforcer) BruteforcePasswordAllUsers(ctx context.Context) ([]scanner.ScanResult, error) {
	err := br.initialize(ctx)
	if err != nil {
		return nil, err
	}

	for _, user := range br.users {
		pass, err := br.bruteforcePasswordsUser(ctx, user)
		if err != nil {
			return nil, err
		}

		err = br.savePasswordHash(ctx, user, pass)
		if err != nil {
			return nil, err
		}

		if pass != "" {
			username, err := user.GetUsername()
			if err != nil {
				return nil, err
			}
			br.results = append(br.results, &bruteforceResult{
				user:     username,
				password: pass,
			})
		}
	}

	return br.results, nil
}

func (br *bruteforcer) markStatusAsSolved(ctx context.Context, user scanner.User, password string) error {
	br.statusLock.Lock()
	defer br.statusLock.Unlock()
	entry, ok := br.status[user]
	if !ok {
		return errors.New("user not found")
	}
	entry.Tried = entry.Total
	entry.FoundPassword = password
	br.status[user] = entry
	return nil
}

func (br *bruteforcer) markIncreaseTried(ctx context.Context, user scanner.User, internalID int64) error {
	br.statusLock.Lock()
	defer br.statusLock.Unlock()
	entry, ok := br.status[user]
	if !ok {
		return errors.New("user not found")
	}
	entry.Tried += 1
	if internalID > entry.MaximumInternalID {
		entry.MaximumInternalID = internalID
	}
	br.status[user] = entry
	return nil
}

func (br *bruteforcer) markStatusAsUnsolved(ctx context.Context, user scanner.User) error {
	br.statusLock.Lock()
	defer br.statusLock.Unlock()
	entry, ok := br.status[user]
	if !ok {
		return errors.New("user not found")
	}
	entry.FoundPassword = ""
	entry.Tried = entry.Total
	br.status[user] = entry
	return nil
}

func (br *bruteforcer) tryPlaintextPassword(ctx context.Context, user scanner.User) (string, error) {
	pass, hasrpw, err := user.GetRawPassword()
	if err != nil {
		return "", err
	}
	if hasrpw {
		exists, err := br.passwordProvider.GetSpecificPassword(pass)
		if err != nil {
			return "", err
		}

		if exists {
			err = br.markStatusAsSolved(ctx, user, pass)
			if err != nil {
				return "", err
			}
		}

		return pass, nil
	}

	return "", nil
}

func (br *bruteforcer) bruteforcePasswordsUser(
	ctx context.Context,
	u scanner.User,
) (string, error) {
	pass, err := br.tryPlaintextPassword(ctx, u)
	if err != nil {
		return "", err
	}
	if pass != "" {
		return pass, nil
	}

	hash, err := u.GetHashedPassword()
	if err != nil {
		return "", err
	}
	username, err := u.GetUsername()
	if err != nil {
		return "", err
	}

	var startBruteforceID int64 = 0

	if hash != "" {
		alreadySolved, lastID, err := br.passwordProvider.GetPasswordByHash(username, hash)
		if err != nil {
			return "", err
		}
		if alreadySolved != "" {
			err = br.markStatusAsSolved(ctx, u, alreadySolved)
			if err != nil {
				return "", err
			}
			return alreadySolved, nil
		}

		startBruteforceID = lastID
	}

	err = br.passwordProvider.Start(startBruteforceID)
	if err != nil {
		return "", err
	}
	defer br.passwordProvider.Close()

	ticker := time.NewTicker(1 * time.Second)
	errorChan := make(chan error, 1)
	resultChan := make(chan string, 1)

	sm := semaphore.NewWeighted(10)
	wg := sync.WaitGroup{}

	var maximumID int64 = 0

	for br.passwordProvider.Next() {
		if err := br.passwordProvider.Error(); err != nil {
			return "", err
		}
		internalID, pass, err := br.passwordProvider.Current()
		if err != nil {
			return "", err
		}
		if internalID > maximumID {
			maximumID = internalID
		}

		sm.Acquire(ctx, 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer sm.Release(1)

			ok, err := u.VerifyPassword(pass)
			if err != nil {
				errorChan <- err
			}

			if ok {
				br.markStatusAsSolved(ctx, u, pass)
				resultChan <- pass
				return
			}

			br.markIncreaseTried(ctx, u, internalID)
		}()

		select {
		case <-ticker.C:
			br.updateStatus()
		case err := <-errorChan:
			return "", err
		case pass := <-resultChan:
			return pass, nil
		default:
		}
	}
	if err := br.passwordProvider.Error(); err != nil {
		return "", err
	}

	wg.Wait()

	br.markStatusAsUnsolved(ctx, u)

	select {
	case err := <-errorChan:
		return "", err
	case pass := <-resultChan:
		return pass, nil
	default:
	}
	br.updateStatus()

	return "", err
}
