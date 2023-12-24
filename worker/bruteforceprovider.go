package worker

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/bruteforce"
	"github.com/tedyst/licenta/scanner"
)

type internalPassword struct {
	Password string `json:"password"`
	ID       int64  `json:"id"`
}

type remoteBruteforceProvider struct {
	client generated.ClientWithResponsesInterface
	task   Task
}

type remotePasswordProvider struct {
	client generated.ClientWithResponsesInterface
	task   Task

	context context.Context

	hasNext      bool
	count        int
	currentBatch []internalPassword
}

func (p *remotePasswordProvider) readBatch() error {
	if len(p.currentBatch) > 1 || !p.hasNext {
		return nil
	}

	lastID := int32(-1)
	if len(p.currentBatch) > 0 {
		lastID = int32(p.currentBatch[len(p.currentBatch)-1].ID)
	}
	response, err := p.client.GetProjectProjectidBruteforcePasswordsWithResponse(p.context, p.task.PostgresScan.Database.ID, &generated.GetProjectProjectidBruteforcePasswordsParams{
		LastId: &lastID,
	})

	if err != nil {
		return err
	}

	switch response.StatusCode() {
	case http.StatusOK:
		for _, password := range response.JSON200.Results {
			p.currentBatch = append(p.currentBatch, internalPassword{
				Password: password.Password,
				ID:       password.Id,
			})
		}
		p.count = response.JSON200.Count
		p.hasNext = response.JSON200.Next != nil
		return nil
	default:
		return errors.New("error getting passwords")
	}
}

func (p *remotePasswordProvider) GetCount() (int, error) {
	err := p.readBatch()
	if err != nil {
		return 0, err
	}

	return p.count, nil
}

func (p *remotePasswordProvider) GetSpecificPassword(password string) (int64, bool, error) {
	return 0, false, nil
}

func (p *remotePasswordProvider) Next() bool {
	err := p.readBatch()
	if err != nil {
		return false
	}

	p.currentBatch = p.currentBatch[1:]

	return len(p.currentBatch) != 0 || p.hasNext
}

func (p *remotePasswordProvider) Error() error {
	return nil
}

func (p *remotePasswordProvider) Current() (int64, string, error) {
	return p.currentBatch[0].ID, p.currentBatch[0].Password, nil
}

func (p *remotePasswordProvider) Start(index int64) error {
	return nil
}

func (p *remotePasswordProvider) Close() {

}

func (p *remotePasswordProvider) SavePasswordHash(username, hash, password string, maxInternalID int64) error {
	return nil
}

func (p *remotePasswordProvider) GetPasswordByHash(username, hash string) (string, int64, error) {
	return "", 0, nil
}

func (p *remoteBruteforceProvider) NewBruteforcer(ctx context.Context, sc scanner.Scanner, statusFunc bruteforce.StatusFunc, projectID int) (bruteforce.Bruteforcer, error) {
	return bruteforce.NewBruteforcer(&remotePasswordProvider{
		client:  p.client,
		task:    p.task,
		context: ctx,
	}, sc, statusFunc), nil
}

var _ bruteforce.BruteforceProvider = (*remoteBruteforceProvider)(nil)
