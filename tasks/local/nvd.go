package local

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/nvd"
)

type nvdRunner struct {
	queries db.TransactionQuerier
}

func NewNVDRunner(queries db.TransactionQuerier) *nvdRunner {
	return &nvdRunner{
		queries: queries,
	}
}

func (r *nvdRunner) importCpesInDB(ctx context.Context, product nvd.Product, database db.TransactionQuerier, result nvd.NvdCpeAPIResult, dbCpes []*queries.NvdCpe) error {
	for _, result := range result.Products {
		var cpe *models.NvdCPE

		version, err := nvd.ExtractCpeVersionProduct(product, result.Cpe.Titles)
		if err != nil {
			continue
		}
		t, err := result.Cpe.LastModifiedDate()
		if err != nil {
			return err
		}

		found := false
		for _, dbCpe := range dbCpes {
			if dbCpe.Cpe == result.Cpe.CpeName {
				cpe = dbCpe
				found = true
				break
			}
		}

		var forceUpdate = false

		if !found {
			slog.DebugContext(ctx, "Creating CPE", slog.Int("product", int(product)), slog.String("cpe", result.Cpe.CpeName))
			cpe, err = database.CreateNvdCPE(ctx, queries.CreateNvdCPEParams{
				Cpe:          result.Cpe.CpeName,
				DatabaseType: int32(product),
				LastModified: pgtype.Timestamptz{Time: t, Valid: true},
				Version:      version,
			})
			if err != nil {
				return err
			}
			forceUpdate = true
		}

		if !cpe.LastModified.Time.Equal(t) || forceUpdate {
			err = database.UpdateNvdCPE(ctx, queries.UpdateNvdCPEParams{
				ID:           cpe.ID,
				LastModified: pgtype.Timestamptz{Time: t, Valid: true},
				Version:      sql.NullString{String: version, Valid: true},
			})
			if err != nil {
				return err
			}

			time.Sleep(6 * time.Second)
			r.updateCVEsForSpecificCPE(ctx, database, product, cpe)
		}
	}

	return nil
}

func (r *nvdRunner) importCVEsInDB(ctx context.Context, product nvd.Product, database db.TransactionQuerier, result nvd.NvdCveAPIResult, cpe *models.NvdCPE) error {
	for _, result := range result.Vulnerabilities {
		var cve *models.NvdCVE
		cve, err := database.GetCveByCveID(ctx, result.Cve.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return errors.Wrap(err, "failed to get cve")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			slog.DebugContext(ctx, "Creating CVE", slog.Int("product", int(product)), slog.String("cpe", cpe.Cpe), slog.String("cve", result.Cve.ID))

			publishedDate, err := result.Cve.PubslihedDate()
			if err != nil {
				return errors.Wrap(err, "failed to parse published date")
			}
			lastModified, err := result.Cve.LastModifiedDate()
			if err != nil {
				return errors.Wrap(err, "failed to parse last modified date")
			}
			score, err := result.Cve.Score()
			if err != nil {
				return errors.Wrap(err, "failed to get score")
			}

			cve, err = database.CreateNvdCve(ctx, queries.CreateNvdCveParams{
				CveID:        result.Cve.ID,
				Description:  result.Cve.Descriptions[0].Value,
				Published:    pgtype.Timestamptz{Time: publishedDate, Valid: true},
				LastModified: pgtype.Timestamptz{Time: lastModified, Valid: true},
				Score:        score,
			})
			if err != nil {
				return errors.Wrap(err, "failed to create cve")
			}
		}

		_, err = database.GetCveCpeByCveAndCpe(ctx, queries.GetCveCpeByCveAndCpeParams{
			CveID: cve.ID,
			CpeID: cpe.ID,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return errors.Wrap(err, "failed to get cvecpe")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = database.CreateNvdCveCPE(ctx, queries.CreateNvdCveCPEParams{
				CveID: cve.ID,
				CpeID: cpe.ID,
			})
			if err != nil {
				return errors.Wrap(err, "failed to create cvecpe")
			}
		}
	}

	return nil
}

func (r *nvdRunner) updateCVEsForSpecificCPE(ctx context.Context, database db.TransactionQuerier, product nvd.Product, cpe *queries.NvdCpe) error {
	var startIndex int64 = 0

	for {
		var reader io.ReadCloser
		reader, err := nvd.DownloadCVEsNext(ctx, product, cpe.Cpe, startIndex)
		if err != nil && !errors.Is(err, nvd.ErrRateLimit) {
			return err
		}
		defer reader.Close()

		if errors.Is(err, nvd.ErrRateLimit) {
			slog.DebugContext(ctx, "Rate limit reached for getting CVEs, waiting 10 seconds", slog.Int("product", int(product)), slog.String("cpe", cpe.Cpe))
			time.Sleep(10 * time.Second)
			continue
		}

		result, err := nvd.ParseCveAPI(ctx, reader)
		if err != nil {
			return err
		}

		err = r.importCVEsInDB(ctx, product, database, result, cpe)
		if err != nil {
			return err
		}

		if result.StartIndex+result.ResultsPerPage < result.TotalResults {
			startIndex += result.ResultsPerPage
		} else {
			break
		}

		slog.DebugContext(ctx, "Waiting 6 seconds before next request", slog.Int("product", int(product)), slog.String("cpe", cpe.Cpe))
		time.Sleep(6 * time.Second)
	}

	return nil
}

func (r *nvdRunner) UpdateNVDVulnerabilitiesForProduct(ctx context.Context, product nvd.Product) (err error) {
	database, err := r.queries.StartTransaction(ctx)
	if err != nil {
		return err
	}
	defer database.EndTransaction(ctx, err)

	var startIndex int64 = 0

	for {
		var reader io.ReadCloser
		reader, err = nvd.DownloadCpeNext(ctx, product, startIndex)
		if err != nil && !errors.Is(err, nvd.ErrRateLimit) {
			return err
		}
		defer reader.Close()

		if errors.Is(err, nvd.ErrRateLimit) {
			slog.DebugContext(ctx, "Rate limit reached for getting CPEs, waiting 10 seconds", slog.Int("product", int(product)))
			time.Sleep(10 * time.Second)
			continue
		}

		result, err := nvd.ParseCpeAPI(ctx, reader)
		if err != nil {
			return err
		}

		dbCpes, err := r.queries.GetNvdCPEsByDBType(ctx, int32(product))
		if err != nil {
			return err
		}

		err = r.importCpesInDB(ctx, product, database, result, dbCpes)
		if err != nil {
			return err
		}

		if result.StartIndex+result.ResultsPerPage < result.TotalResults {
			startIndex += result.ResultsPerPage
		} else {
			break
		}

		slog.DebugContext(ctx, "Waiting 6 seconds before next request", slog.Int("product", int(product)))
		time.Sleep(6 * time.Second)
	}
	return nil
}
