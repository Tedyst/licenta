package handlers

import (
	"context"

	"github.com/tedyst/licenta/api/v1/generated"
	. "github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/nvd"
)

func (server *serverHandler) GetCvesDbTypeVersion(ctx context.Context, request GetCvesDbTypeVersionRequestObject) (GetCvesDbTypeVersionResponseObject, error) {
	worker, err := server.workerauth.GetWorker(ctx)
	if err != nil {
		return nil, err
	}
	if worker == nil {
		return GetCvesDbTypeVersion401JSONResponse{
			Message: "Unauthorized",
			Success: false,
		}, nil
	}

	product, err := nvd.GetNvdDatabaseType(request.DbType)
	if err != nil {
		return GetCvesDbTypeVersion404JSONResponse{
			Message: "Database type not found",
			Success: false,
		}, nil
	}

	cves, err := server.DatabaseProvider.GetCvesByProductAndVersion(ctx, queries.GetCvesByProductAndVersionParams{
		DatabaseType: int32(product),
		Version:      request.Version,
	})
	if err != nil {
		return nil, err
	}

	var result []generated.CVE
	for _, cve := range cves {
		result = append(result, generated.CVE{
			CveId:        cve.NvdCfe.CveID,
			Description:  cve.NvdCfe.Description,
			Id:           int64(cve.NvdCfe.ID),
			LastModified: cve.NvdCfe.LastModified.Time.Format("2006-01-02T15:04:05Z"),
			PublishedAt:  cve.NvdCfe.Published.Time.Format("2006-01-02T15:04:05Z"),
		})
	}

	return GetCvesDbTypeVersion200JSONResponse{
		Success: true,
		Cves:    result,
	}, nil
}
