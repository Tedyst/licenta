package local

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"

	"errors"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/tedyst/licenta/db/queries"
	"github.com/tedyst/licenta/extractors/docker"
	"github.com/tedyst/licenta/extractors/file"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/scanner"
)

type DockerRunner struct {
	queries DockerQuerier

	FileScannerProvider   func(opts ...file.Option) (*file.FileScanner, error)
	DockerScannerProvider func(ctx context.Context, fileScanner docker.FileScanner, imageName string, opts ...docker.Option) (*docker.DockerScan, error)
	saltKey               string
}

type DockerQuerier interface {
	GetDockerScannedLayersForImage(ctx context.Context, imageID int64) ([]string, error)
	CreateDockerScannedLayerForProject(ctx context.Context, params queries.CreateDockerScannedLayerForProjectParams) (*queries.DockerLayer, error)
	CreateDockerLayerResultsForProject(ctx context.Context, params []queries.CreateDockerLayerResultsForProjectParams) (int64, error)
	CreateScanResult(ctx context.Context, params queries.CreateScanResultParams) (*queries.ScanResult, error)
}

func NewDockerRunner(queries DockerQuerier, saltKey string) *DockerRunner {
	return &DockerRunner{
		queries:               queries,
		FileScannerProvider:   file.NewScanner,
		DockerScannerProvider: docker.NewScanner,
		saltKey:               saltKey,
	}
}

func (r *DockerRunner) ScanDockerRepository(ctx context.Context, image *queries.DockerImage, scan *queries.Scan) (err error) {
	if image == nil {
		return errors.New("image is nil")
	}

	scannnedMap := map[string]bool{}
	mutex := sync.Mutex{}
	alreadyCreated := map[string]*queries.DockerLayer{}

	resultCallback := func(scc *docker.DockerScan, result *docker.LayerResult) error {
		mutex.Lock()
		defer mutex.Unlock()
		if err != nil {
			return fmt.Errorf("ScanDockerRepository: cannot update layer scan: %w", err)
		}

		var scannedLayer *queries.DockerLayer
		if layer, ok := alreadyCreated[result.Layer]; ok {
			scannedLayer = layer
		} else {
			scannedLayer, err = r.queries.CreateDockerScannedLayerForProject(ctx, queries.CreateDockerScannedLayerForProjectParams{
				LayerHash: result.Layer,
				ImageID:   image.ID,
			})
			if err != nil {
				return fmt.Errorf("ScanDockerRepository: cannot create scanned layer: %w", err)
			}
			alreadyCreated[result.Layer] = scannedLayer
		}

		layerResults := []queries.CreateDockerLayerResultsForProjectParams{}
		for _, fileResult := range result.Results {
			layerResults = append(layerResults, queries.CreateDockerLayerResultsForProjectParams{
				LayerID:       scannedLayer.ID,
				Name:          fileResult.Name,
				Line:          fileResult.Line,
				LineNumber:    int32(fileResult.LineNumber),
				Match:         fileResult.Match,
				Probability:   fileResult.Probability,
				Username:      sql.NullString{String: fileResult.Username, Valid: fileResult.Username != ""},
				Password:      sql.NullString{String: fileResult.Password, Valid: fileResult.Password != ""},
				Filename:      fileResult.FileName,
				PreviousLines: fileResult.PreviousLines,
			})

			_, err = r.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
				ScanID:     scan.ID,
				Severity:   int32(scanner.SEVERITY_MEDIUM),
				Message:    "Found " + fileResult.Match + " in " + fileResult.Line,
				ScanSource: models.SCAN_DOCKER,
			})
			if err != nil {
				return fmt.Errorf("ScanDockerRepository: cannot create scan result: %w", err)
			}
		}

		if len(layerResults) == 0 {
			return nil
		}

		count, err := r.queries.CreateDockerLayerResultsForProject(ctx, layerResults)
		if err != nil {
			return fmt.Errorf("ScanDockerRepository: cannot create layer results: %w", err)
		}

		slog.DebugContext(ctx, "ScanDockerRepository: created layer results", "count", count)

		return nil
	}

	options := []docker.Option{}
	options = append(options, docker.WithCallbackResult(resultCallback))
	if image.Username != "" && image.Password != "" {
		auth := authn.Basic{
			Username: image.Username,
			Password: image.Password,
		}
		options = append(options, docker.WithCredentials(&auth))
	}

	if image.MinProbability.Valid {
		options = append(options, docker.WithProbability(image.MinProbability.Float64))
	}

	fileOptions := []file.Option{}

	if image.ProbabilityIncreaseMultiplier.Valid {
		fileOptions = append(fileOptions, file.WithProbabilityIncreaseMultiplier(image.ProbabilityIncreaseMultiplier.Float64))
	}
	if image.ProbabilityDecreaseMultiplier.Valid {
		fileOptions = append(fileOptions, file.WithProbabilityDecreaseMultiplier(image.ProbabilityDecreaseMultiplier.Float64))
	}
	if image.EntropyThreshold.Valid {
		fileOptions = append(fileOptions, file.WithEntropyThresholdMidpoint(int(image.EntropyThreshold.Float64)))
	}
	if image.LogisticGrowthRate.Valid {
		fileOptions = append(fileOptions, file.WithLogisticGrowthRate(image.LogisticGrowthRate.Float64))
	}

	fs, err := r.FileScannerProvider(fileOptions...)
	if err != nil {
		return fmt.Errorf("ScanDockerRepository: cannot create file scanner: %w", err)
	}

	scannedLayers, err := r.queries.GetDockerScannedLayersForImage(ctx, image.ID)
	if err != nil {
		return err
	}

	for _, layer := range scannedLayers {
		scannnedMap[layer] = true
	}

	scc, err := r.DockerScannerProvider(ctx, fs, image.DockerImage, options...)
	if err != nil {
		return err
	}

	layers, err := scc.FindLayers(ctx)
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

	for _, layer := range scanLayers {
		digest, err := layer.Digest()
		if err != nil {
			return err
		}

		_, err = r.queries.CreateScanResult(ctx, queries.CreateScanResultParams{
			ScanID:     scan.ID,
			Severity:   int32(scanner.SEVERITY_INFORMATIONAL),
			Message:    "Started scanning the layer " + digest.String(),
			ScanSource: models.SCAN_DOCKER,
		})
		if err != nil {
			return fmt.Errorf("ScanDockerRepository: cannot create scan result: %w", err)
		}
	}

	err = scc.ProcessLayers(ctx, scanLayers)
	if err != nil {
		return err
	}

	return nil
}
