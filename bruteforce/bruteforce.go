package bruteforce

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/tedyst/licenta/scanner"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	updateStatus     func(ctx context.Context) error

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
	br.updateStatus = func(ctx context.Context) error {
		br.statusLock.Lock()
		defer br.statusLock.Unlock()

		for user, value := range br.status {
			username, err := user.GetUsername()
			if err != nil {
				return err
			}
			slog.DebugContext(ctx, "Updating bruteforce status for user", "user", username, "tried", value.Tried, "total", value.Total, "found_password", value.FoundPassword, "maximum_internal_id", value.MaximumInternalID)
		}

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

	return br.updateStatus(ctx)
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
	ctx, span := tracer.Start(ctx, "BruteforcePasswordAllUsers")
	defer span.End()

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

		username, err := user.GetUsername()
		if err != nil {
			return nil, err
		}
		br.results = append(br.results, &bruteforceResult{
			user:     username,
			password: pass,
		})
	}

	return br.results, nil
}

func (br *bruteforcer) markStatusAsSolved(ctx context.Context, user scanner.User, password string, internalID int64) error {
	br.statusLock.Lock()
	defer br.statusLock.Unlock()
	entry, ok := br.status[user]
	if !ok {
		return errors.New("user not found")
	}
	entry.Tried = entry.Total
	entry.FoundPassword = password
	entry.MaximumInternalID = internalID
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

func (br *bruteforcer) setMaximumInternalID(ctx context.Context, user scanner.User, internalID int64) error {
	br.statusLock.Lock()
	defer br.statusLock.Unlock()
	entry, ok := br.status[user]
	if !ok {
		return errors.New("user not found")
	}
	entry.MaximumInternalID = internalID
	br.status[user] = entry
	return nil
}

func (br *bruteforcer) tryPlaintextPassword(ctx context.Context, user scanner.User) (string, error) {
	pass, hasrpw, err := user.GetRawPassword()
	if err != nil {
		return "", err
	}
	if hasrpw {
		internalID, exists, err := br.passwordProvider.GetSpecificPassword(pass)
		if err != nil {
			return "", err
		}

		if exists {
			err = br.markStatusAsSolved(ctx, user, pass, internalID)
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
	username, err := u.GetUsername()
	if err != nil {
		return "", err
	}

	ctx, span := tracer.Start(ctx, "bruteforcePasswordsUser", trace.WithAttributes(attribute.KeyValue{
		Key:   attribute.Key("user"),
		Value: attribute.StringValue(username),
	}))
	defer span.End()

	slog.DebugContext(ctx, "Bruteforcing passwords for user", "user", username)

	defer br.updateStatus(ctx)

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

	var startBruteforceID int64 = 0

	if hash != "" {
		alreadySolved, lastID, err := br.passwordProvider.GetPasswordByHash(username, hash)
		if err != nil {
			return "", err
		}
		if alreadySolved != "" {
			slog.DebugContext(ctx, "Password already solved", "user", username, "password", alreadySolved, "hash", hash, "last_id", lastID)
			err = br.markStatusAsSolved(ctx, u, alreadySolved, lastID)
			if err != nil {
				return "", err
			}
			passwordsAlreadyFound.Add(ctx, 1)
			return alreadySolved, nil
		}

		slog.DebugContext(ctx, "Password not already solved, but continuing from last_id", "user", username, "hash", hash, "last_id", lastID)

		startBruteforceID = lastID
	}

	br.setMaximumInternalID(ctx, u, startBruteforceID)

	err = br.passwordProvider.Start(startBruteforceID)
	if err != nil {
		return "", err
	}
	defer br.passwordProvider.Close()

	ticker := time.NewTicker(1 * time.Second)
	errorChan := make(chan error, 1)
	resultChan := make(chan struct {
		password   string
		internalID int64
	}, 1)

	sm := semaphore.NewWeighted(10)
	wg := sync.WaitGroup{}

	for br.passwordProvider.Next() {
		if err := br.passwordProvider.Error(); err != nil {
			return "", err
		}
		internalID, pass, err := br.passwordProvider.Current()
		if err != nil {
			return "", err
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

			passwordsTried.Add(ctx, 1)

			if ok {
				br.markStatusAsSolved(ctx, u, pass, internalID)
				resultChan <- struct {
					password   string
					internalID int64
				}{pass, internalID}
				return
			}

			br.markIncreaseTried(ctx, u, internalID)
		}()

		select {
		case <-ticker.C:
			br.updateStatus(ctx)
		case err := <-errorChan:
			return "", err
		case pass := <-resultChan:
			return pass.password, nil
		default:
		}
	}
	if err := br.passwordProvider.Error(); err != nil {
		return "", err
	}

	wg.Wait()

	select {
	case err := <-errorChan:
		br.markStatusAsUnsolved(ctx, u)
		return "", err
	case pass := <-resultChan:
		br.markStatusAsSolved(ctx, u, pass.password, pass.internalID)
		return pass.password, nil
	default:
		br.markStatusAsUnsolved(ctx, u)
		return "", err
	}
}
