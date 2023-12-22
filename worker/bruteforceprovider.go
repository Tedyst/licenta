package worker

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

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
	remoteURL string
	authToken string
	task      Task
}

type remotePasswordProvider struct {
	remoteURL string
	authToken string
	task      Task

	context context.Context

	nextURL      string
	count        int
	currentBatch []internalPassword
}

func (p *remotePasswordProvider) readBatch() error {
	if len(p.currentBatch) > 1 {
		return nil
	}

	var remoteURL string
	if p.nextURL != "" {
		remoteURL = p.remoteURL + p.nextURL
	} else {
		remoteURL = p.remoteURL + "/api/v1/project/" + strconv.Itoa(int(p.task.PostgresScan.Database.ID)) + "/bruteforce-passwords/"
	}

	req, err := http.NewRequest("GET", remoteURL, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(p.context)
	req.Header.Set("X-Worker-Token", p.authToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("error getting passwords")
	}

	var response generated.GetProjectProjectidBruteforcePasswords200JSONResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	if !response.Success {
		return errors.New("error getting passwords")
	}

	p.count = response.Count
	if response.Next == nil {
		p.nextURL = ""
	} else {
		p.nextURL = *response.Next
	}

	for _, password := range response.Results {
		p.currentBatch = append(p.currentBatch, internalPassword{
			Password: password.Password,
			ID:       password.Id,
		})
	}

	return nil
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

	return len(p.currentBatch) != 0 && p.nextURL != ""
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
		remoteURL: p.remoteURL,
		authToken: p.authToken,
		task:      p.task,
		context:   ctx,
	}, sc, statusFunc), nil
}

var _ bruteforce.BruteforceProvider = (*remoteBruteforceProvider)(nil)
