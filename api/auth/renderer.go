package auth

import (
	"context"
	"encoding/json"

	"github.com/volatiletech/authboss/v3"
)

var (
	jsonDefaultFailures = []string{authboss.DataErr, authboss.DataValidation}
)

type jsonRenderer struct {
	Failures []string
}

func (jsonRenderer) Load(names ...string) error {
	return nil
}

func (j jsonRenderer) Render(ctx context.Context, page string, data authboss.HTMLData) (output []byte, contentType string, err error) {
	if data == nil {
		return []byte(`{"success":true}`), "application/json", nil
	}

	if _, hasStatus := data["status"]; !hasStatus {
		failures := j.Failures
		if len(failures) == 0 {
			failures = jsonDefaultFailures
		}

		success := true
		for _, failure := range failures {
			val, has := data[failure]
			if has && val != nil {
				success = false
				break
			}
		}

		data["success"] = success
	} else {
		data["success"] = data["status"] == "success"
		delete(data, "status")
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, "", err
	}

	return b, "application/json", nil
}
