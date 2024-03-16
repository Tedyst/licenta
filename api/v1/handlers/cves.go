package handlers

import (
	"context"
	"time"

	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/nvd"
)

func (server *serverHandler) GetCvesDbTypeVersion(ctx context.Context, request generated.GetCvesDbTypeVersionRequestObject) (generated.GetCvesDbTypeVersionResponseObject, error) {
	worker, err := server.workerauth.GetWorker(ctx)
	if err != nil {
		return nil, err
	}
	if worker == nil {
		return generated.GetCvesDbTypeVersion401JSONResponse{
			Message: "Unauthorized",
			Success: false,
		}, nil
	}

	cves, err := server.DatabaseProvider.GetCvesByProductAndVersion(ctx, queries.GetCvesByProductAndVersionParams{
		DatabaseType: int32(nvd.GetNvdProductType(request.DbType)),
		Version:      request.Version,
	})
	if err != nil {
		return nil, err
	}

	result := make([]generated.CVE, 0, len(cves))
	for _, cve := range cves {
		result = append(result, generated.CVE{
			CveId:        cve.NvdCfe.CveID,
			Description:  cve.NvdCfe.Description,
			Id:           int64(cve.NvdCfe.ID),
			LastModified: cve.NvdCfe.LastModified.Time.Format(time.RFC3339Nano),
			PublishedAt:  cve.NvdCfe.Published.Time.Format(time.RFC3339Nano),
		})
	}

	return generated.GetCvesDbTypeVersion200JSONResponse{
		Success: true,
		Cves:    result,
	}, nil
}
