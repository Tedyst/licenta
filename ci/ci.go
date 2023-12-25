package ci

import (
	"context"

	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/models"
)

func SignalFullProjectScan(ctx context.Context, client generated.ClientWithResponsesInterface, project *models.Project) error {
	return nil
}
