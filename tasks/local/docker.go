package local

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"

	"github.com/pkg/errors"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/extractors/docker"
	"github.com/tedyst/licenta/extractors/file"
	"github.com/tedyst/licenta/models"
)

type dockerRunner struct {
	queries db.TransactionQuerier
}

func NewDockerRunner(queries db.TransactionQuerier) *dockerRunner {
	return &dockerRunner{
		queries: queries,
	}
}

func (r *dockerRunner) ScanDockerRepository(ctx context.Context, image *models.ProjectDockerImage) (err error) {
	if image == nil {
		return errors.New("image is nil")
	}

	querier, err := r.queries.StartTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "ScanDockerRepository: cannot start transaction")
	}
	defer func() {
		err = querier.EndTransaction(ctx, err)
	}()

	scannnedMap := map[string]bool{}
	mutex := sync.Mutex{}

	resultCallback := func(scanner *docker.DockerScan, result *docker.LayerResult) error {
		mutex.Lock()
		defer mutex.Unlock()
		finished := false
		if len(scannnedMap) == len(scanner.ScannedLayers()) {
			finished = true
		}
		querier.UpdateDockerLayerScanForProject(ctx, queries.UpdateDockerLayerScanForProjectParams{
			ProjectID:     image.ProjectID,
			DockerImage:   image.ID,
			ScannedLayers: int32(len(scanner.ScannedLayers())),
			Finished:      finished,
		})
		scannedLayer, err := querier.CreateDockerScannedLayerForProject(ctx, queries.CreateDockerScannedLayerForProjectParams{
			ProjectID: image.ProjectID,
			LayerHash: result.Layer,
		})
		if err != nil {
			return errors.Wrap(err, "ScanDockerRepository: cannot create scanned layer")
		}

		layerResults := []queries.CreateDockerLayerResultsForProjectParams{}
		for _, fileResult := range result.Results {
			layerResults = append(layerResults, queries.CreateDockerLayerResultsForProjectParams{
				ProjectID:   image.ProjectID,
				Layer:       scannedLayer.ID,
				Name:        fileResult.Name,
				Line:        fileResult.Line,
				LineNumber:  int32(fileResult.LineNumber),
				Match:       fileResult.Match,
				Probability: fileResult.Probability,
				Username:    sql.NullString{String: fileResult.Username, Valid: fileResult.Username != ""},
				Password:    sql.NullString{String: fileResult.Password, Valid: fileResult.Password != ""},
				Filename:    fileResult.FileName,
			})
		}

		count, err := querier.CreateDockerLayerResultsForProject(ctx, layerResults)
		if err != nil {
			return errors.Wrap(err, "ScanDockerRepository: cannot create layer results")
		}

		slog.DebugContext(ctx, "ScanDockerRepository: created layer results", "count", count)

		return nil
	}

	options := []docker.Option{}
	options = append(options, docker.WithCallbackResult(resultCallback))
	if image.Username.Valid && image.Password.Valid {
		auth := authn.Basic{
			Username: image.Username.String,
			Password: image.Password.String,
		}
		options = append(options, docker.WithCredentials(&auth))
	}

	if image.MinProbability.Valid {
		options = append(options, docker.WithProbability(image.MinProbability.Float64))
	}

	fileOptions := []file.Option{}
	if image.UseDefaultPasswordsCompletelyIgnore {
		return errors.New("UseDefaultPasswordsCompletelyIgnore is not supported")
	}
	if image.UseDefaultUsernamesCompletelyIgnore {
		return errors.New("UseDefaultUsernamesCompletelyIgnore is not supported")
	}
	if image.UseDefaultWordsIncreaseProbability {
		return errors.New("UseDefaultWordsIncreaseProbability is not supported")
	}
	if image.UseDefaultWordsReduceProbability {
		return errors.New("UseDefaultWordsReduceProbability is not supported")
	}

	if image.ProbabilityIncreaseMultiplier.Valid {
		fileOptions = append(fileOptions, file.WithProbabilityIncreaseMultiplier(image.ProbabilityIncreaseMultiplier.Float64))
	}
	if image.ProbailityDecreaseMultiplier.Valid {
		fileOptions = append(fileOptions, file.WithProbabilityDecreaseMultiplier(image.ProbailityDecreaseMultiplier.Float64))
	}
	if image.EntropyThreshold.Valid {
		fileOptions = append(fileOptions, file.WithEntropyThresholdMidpoint(int(image.EntropyThreshold.Int32)))
	}
	if image.LogisticGrowthRate.Valid {
		fileOptions = append(fileOptions, file.WithLogisticGrowthRate(image.LogisticGrowthRate.Float64))
	}

	if len(fileOptions) > 0 {
		options = append(options, docker.WithFileScannerOptions(fileOptions...))
	}

	scannedLayers, err := querier.GetDockerScannedLayersForProject(ctx, image.ProjectID)
	if err != nil {
		return err
	}

	for _, layer := range scannedLayers {
		scannnedMap[layer] = true
	}

	scanner, err := docker.NewScanner(ctx, image.DockerImage, options...)
	if err != nil {
		return err
	}

	layers, err := scanner.FindLayers(ctx)
	if err != nil {
		return err
	}

	var scanLayers []v1.Layer
	for _, layer := range layers {
		digest, err := layer.Digest()
		if err != nil {
			return err
		}
		if _, ok := scannnedMap[digest.String()]; !ok {
			scanLayers = append(scanLayers, layer)
		}
	}

	err = scanner.ProcessLayers(ctx, scanLayers)
	if err != nil {
		return err
	}

	return nil
}
