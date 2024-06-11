package local

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"time"

	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/nvd"
)

type NvdQuerier interface {
	GetNvdCPEsByDBType(ctx context.Context, databaseType int32) ([]*queries.NvdCpe, error)
	CreateNvdCPE(ctx context.Context, params queries.CreateNvdCPEParams) (*queries.NvdCpe, error)
	UpdateNvdCPE(ctx context.Context, params queries.UpdateNvdCPEParams) error
	GetCveByCveID(ctx context.Context, cveID string) (*queries.NvdCfe, error)
	CreateNvdCve(ctx context.Context, params queries.CreateNvdCveParams) (*queries.NvdCfe, error)
	GetCveCpeByCveAndCpe(ctx context.Context, params queries.GetCveCpeByCveAndCpeParams) (*queries.NvdCveCpe, error)
	CreateNvdCveCPE(ctx context.Context, params queries.CreateNvdCveCPEParams) (*queries.NvdCveCpe, error)
}

type NvdRunner struct {
	queries NvdQuerier
}

func NewNVDRunner(queries NvdQuerier) *NvdRunner {
	return &NvdRunner{
		queries: queries,
	}
}

func (r *NvdRunner) importCpesInDB(ctx context.Context, product nvd.Product, database NvdQuerier, result nvd.NvdCpeAPIResult, dbCpes []*queries.NvdCpe) error {
	ctx, span := tracer.Start(ctx, "importCpesInDB")
	defer span.End()

	for _, result := range result.Products {
		var cpe *queries.NvdCpe

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
			slog.DebugContext(ctx, "Created CPE", slog.Int("product", int(product)), slog.String("cpe", result.Cpe.CpeName))
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

			slog.DebugContext(ctx, "Waiting 6 seconds before next request", slog.Int("product", int(product)), slog.String("cpe", cpe.Cpe))

			time.Sleep(6 * time.Second)
			err := r.updateCVEsForSpecificCPE(ctx, database, product, cpe)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *NvdRunner) importCVEsInDB(ctx context.Context, product nvd.Product, database NvdQuerier, result nvd.NvdCveAPIResult, cpe *queries.NvdCpe) error {
	ctx, span := tracer.Start(ctx, "importCVEsInDB")
	defer span.End()

	for _, result := range result.Vulnerabilities {
		var cve *queries.NvdCfe
		cve, err := database.GetCveByCveID(ctx, result.Cve.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get cve: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			slog.DebugContext(ctx, "Creating CVE", slog.Int("product", int(product)), slog.String("cpe", cpe.Cpe), slog.String("cve", result.Cve.ID))

			publishedDate, err := result.Cve.PubslihedDate()
			if err != nil {
				return fmt.Errorf("failed to parse published date: %w", err)
			}
			lastModified, err := result.Cve.LastModifiedDate()
			if err != nil {
				return fmt.Errorf("failed to parse last modified date: %w", err)
			}
			score, err := result.Cve.Score()
			if err != nil {
				return fmt.Errorf("failed to get score: %w", err)
			}

			cve, err = database.CreateNvdCve(ctx, queries.CreateNvdCveParams{
				CveID:        result.Cve.ID,
				Description:  result.Cve.Descriptions[0].Value,
				Published:    pgtype.Timestamptz{Time: publishedDate, Valid: true},
				LastModified: pgtype.Timestamptz{Time: lastModified, Valid: true},
				Score:        score,
			})
			if err != nil {
				return fmt.Errorf("failed to create cve: %w", err)
			}

			slog.DebugContext(ctx, "Created CVE", slog.Int("product", int(product)), slog.String("cpe", cpe.Cpe), slog.String("cve", result.Cve.ID))
		}

		_, err = database.GetCveCpeByCveAndCpe(ctx, queries.GetCveCpeByCveAndCpeParams{
			CveID: cve.ID,
			CpeID: cpe.ID,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to get cvecpe: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = database.CreateNvdCveCPE(ctx, queries.CreateNvdCveCPEParams{
				CveID: cve.ID,
				CpeID: cpe.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to create cvecpe: %w", err)
			}
		}
	}

	return nil
}

func (r *NvdRunner) updateCVEsForSpecificCPE(ctx context.Context, database NvdQuerier, product nvd.Product, cpe *queries.NvdCpe) (err error) {
	ctx, span := tracer.Start(ctx, "updateCVEsForSpecificCPE")
	defer span.End()

	var startIndex int64 = 0

	for {
		var reader io.ReadCloser
		reader, err := nvd.DownloadCVEsNext(ctx, product, cpe.Cpe, startIndex)
		if err != nil && !errors.Is(err, nvd.ErrRateLimit) {
			return err
		}
		defer func() {
			err = errors.Join(err, reader.Close())
		}()

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

func (r *NvdRunner) UpdateNVDVulnerabilitiesForProduct(ctx context.Context, product nvd.Product) (err error) {
	ctx, span := tracer.Start(ctx, "UpdateNVDVulnerabilitiesForProduct")
	defer span.End()

	var startIndex int64 = 0

	for {
		var reader io.ReadCloser
		reader, err = nvd.DownloadCpeNext(ctx, product, startIndex)
		if err != nil && !errors.Is(err, nvd.ErrRateLimit) {
			return err
		}
		defer func() {
			err = errors.Join(err, reader.Close())
		}()

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

		err = r.importCpesInDB(ctx, product, r.queries, result, dbCpes)
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
