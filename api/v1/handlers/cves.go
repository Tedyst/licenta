package handlers

import (
	"context"

	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/nvd"
)

func (server *serverHandler) GetCvesDatabaseTypeVersion(ctx context.Context, request generated.GetCvesDatabaseTypeVersionRequestObject) (generated.GetCvesDatabaseTypeVersionResponseObject, error) {
	worker, err := server.workerauth.GetWorker(ctx)
	if err != nil {
		return nil, err
	}
	if worker == nil {
		return generated.GetCvesDatabaseTypeVersion401JSONResponse{
			Message: "Unauthorized",
			Success: false,
		}, nil
	}

	product, err := nvd.GetNvdDatabaseType(request.DatabaseType)
	if err != nil {
		return generated.GetCvesDatabaseTypeVersion404JSONResponse{
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

	return generated.GetCvesDatabaseTypeVersion200JSONResponse{
		Success: true,
		Cves:    result,
	}, nil
}
