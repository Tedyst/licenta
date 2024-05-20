package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/api/authorization"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

func (server *serverHandler) DeleteDockerId(ctx context.Context, request generated.DeleteDockerIdRequestObject) (generated.DeleteDockerIdResponseObject, error) {
	dockerImage, err := server.DatabaseProvider.GetDockerImage(ctx, request.Id)
	if err == pgx.ErrNoRows {
		return generated.DeleteDockerId404JSONResponse{
			Success: false,
			Message: "Docker image not found",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("DeleteDockerId: error getting docker image: %w", err)
	}

	_, _, response, err := checkUserHasProjectPermission[generated.DeleteDockerId401JSONResponse](server, ctx, dockerImage.ProjectID, authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	err = server.DatabaseProvider.DeleteDockerImage(ctx, dockerImage.ID)
	if err != nil {
		return nil, fmt.Errorf("DeleteDockerId: error deleting docker image: %w", err)
	}

	return generated.DeleteDockerId204JSONResponse{
		Success: true,
	}, nil
}

func (server *serverHandler) GetDocker(ctx context.Context, request generated.GetDockerRequestObject) (generated.GetDockerResponseObject, error) {
	dockerImages, err := server.DatabaseProvider.GetDockerImagesForProject(ctx, int64(request.Params.Project))
	if err == pgx.ErrNoRows {
		return generated.GetDocker401JSONResponse{
			Success: false,
			Message: "Docker image not found",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetDocker: error getting docker image: %w", err)
	}

	dockerImagesResponse := make([]generated.DockerImage, len(dockerImages))
	for i, dockerImage := range dockerImages {
		dockerImagesResponse[i] = generated.DockerImage{
			DockerImage:                   dockerImage.DockerImage,
			EntropyThreshold:              float32(dockerImage.EntropyThreshold.Float64),
			LogisticGrowthRate:            float32(dockerImage.LogisticGrowthRate.Float64),
			MinProbability:                float32(dockerImage.MinProbability.Float64),
			Password:                      dockerImage.Password.String,
			ProbabilityDecreaseMultiplier: float32(dockerImage.ProbabilityDecreaseMultiplier.Float64),
			ProbabilityIncreaseMultiplier: float32(dockerImage.ProbabilityIncreaseMultiplier.Float64),
			Id:                            int(dockerImage.ID),
		}
	}

	return generated.GetDocker200JSONResponse{
		Success: true,
		Images:  dockerImagesResponse,
	}, nil
}

func (server *serverHandler) GetDockerId(ctx context.Context, request generated.GetDockerIdRequestObject) (generated.GetDockerIdResponseObject, error) {
	dockerImage, err := server.DatabaseProvider.GetDockerImage(ctx, request.Id)
	if err == pgx.ErrNoRows {
		return generated.GetDockerId404JSONResponse{
			Success: false,
			Message: "Docker image not found",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetDockerId: error getting docker image: %w", err)
	}

	_, _, response, err := checkUserHasProjectPermission[generated.GetDockerId401JSONResponse](server, ctx, dockerImage.ProjectID, authorization.Viewer)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	dbLayers, err := server.DatabaseProvider.GetDockerLayersAndResultsForScan(ctx, dockerImage.ID)
	if err != nil {
		return nil, fmt.Errorf("GetDockerId: error getting docker layer scans: %w", err)
	}

	layers := []generated.DockerLayer{}
	layerResults := map[int64][]generated.DockerLayerResult{}

	for _, dbLayer := range dbLayers {
		if _, ok := layerResults[dbLayer.DockerLayer.ID]; !ok {
			layerResults[dbLayer.DockerLayer.ID] = []generated.DockerLayerResult{}
			layers = append(layers, generated.DockerLayer{
				Id:        int(dbLayer.DockerLayer.ID),
				ProjectId: int(dbLayer.DockerLayer.ProjectID),
				LayerHash: dbLayer.DockerLayer.LayerHash,
				Results:   []generated.DockerLayerResult{},
				ScanId:    int(dbLayer.DockerLayer.ScanID),
			})
		}

		layerResults[dbLayer.DockerLayer.ID] = append(layerResults[dbLayer.DockerLayer.ID], generated.DockerLayerResult{
			CreatedAt:   dbLayer.DockerResult.CreatedAt.Time.Format(time.RFC3339),
			Filename:    dbLayer.DockerResult.Filename,
			Id:          int(dbLayer.DockerResult.ID),
			Layer:       int(dbLayer.DockerResult.LayerID),
			Line:        dbLayer.DockerResult.Line,
			LineNumber:  int(dbLayer.DockerResult.LineNumber),
			Match:       dbLayer.DockerResult.Match,
			Name:        dbLayer.DockerResult.Name,
			Password:    dbLayer.DockerResult.Password.String,
			Probability: float32(dbLayer.DockerResult.Probability),
			ProjectId:   int(dbLayer.DockerResult.ProjectID),
			Username:    dbLayer.DockerResult.Username.String,
		})
	}

	for i, layer := range layers {
		layers[i].Results = layerResults[int64(layer.Id)]
	}

	return generated.GetDockerId200JSONResponse{
		Success: true,
		Image: generated.DockerImage{
			DockerImage:                   dockerImage.DockerImage,
			EntropyThreshold:              float32(dockerImage.EntropyThreshold.Float64),
			LogisticGrowthRate:            float32(dockerImage.LogisticGrowthRate.Float64),
			MinProbability:                float32(dockerImage.MinProbability.Float64),
			Password:                      dockerImage.Password.String,
			ProbabilityDecreaseMultiplier: float32(dockerImage.ProbabilityDecreaseMultiplier.Float64),
			ProbabilityIncreaseMultiplier: float32(dockerImage.ProbabilityIncreaseMultiplier.Float64),
			Id:                            int(dockerImage.ID),
		},
		Layers: layers,
	}, nil
}

func (server *serverHandler) PatchDockerId(ctx context.Context, request generated.PatchDockerIdRequestObject) (generated.PatchDockerIdResponseObject, error) {
	dockerImage, err := server.DatabaseProvider.GetDockerImage(ctx, request.Id)
	if err == pgx.ErrNoRows {
		return generated.PatchDockerId404JSONResponse{
			Success: false,
			Message: "Docker image not found",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("PatchDockerId: error getting docker image: %w", err)
	}

	_, _, response, err := checkUserHasProjectPermission[generated.PatchDockerId401JSONResponse](server, ctx, dockerImage.ProjectID, authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	if request.Body.DockerImage != nil {
		dockerImage.DockerImage = *request.Body.DockerImage
	}
	if request.Body.EntropyThreshold != nil {
		dockerImage.EntropyThreshold = sql.NullFloat64{Float64: float64(*request.Body.EntropyThreshold), Valid: true}
	}
	if request.Body.LogisticGrowthRate != nil {
		dockerImage.LogisticGrowthRate = sql.NullFloat64{Float64: float64(*request.Body.LogisticGrowthRate), Valid: true}
	}
	if request.Body.MinProbability != nil {
		dockerImage.MinProbability = sql.NullFloat64{Float64: float64(*request.Body.MinProbability), Valid: true}
	}
	if request.Body.Password != nil {
		dockerImage.Password = sql.NullString{String: *request.Body.Password, Valid: true}
	}
	if request.Body.ProbabilityDecreaseMultiplier != nil {
		dockerImage.ProbabilityDecreaseMultiplier = sql.NullFloat64{Float64: float64(*request.Body.ProbabilityDecreaseMultiplier), Valid: true}
	}
	if request.Body.ProbabilityIncreaseMultiplier != nil {
		dockerImage.ProbabilityIncreaseMultiplier = sql.NullFloat64{Float64: float64(*request.Body.ProbabilityIncreaseMultiplier), Valid: true}
	}

	image, err := server.DatabaseProvider.UpdateDockerImage(ctx, queries.UpdateDockerImageParams{
		ID:                            dockerImage.ID,
		DockerImage:                   dockerImage.DockerImage,
		Username:                      dockerImage.Username,
		Password:                      dockerImage.Password,
		MinProbability:                dockerImage.MinProbability,
		ProbabilityDecreaseMultiplier: dockerImage.ProbabilityDecreaseMultiplier,
		ProbabilityIncreaseMultiplier: dockerImage.ProbabilityIncreaseMultiplier,
		EntropyThreshold:              dockerImage.EntropyThreshold,
		LogisticGrowthRate:            dockerImage.LogisticGrowthRate,
	})
	if err != nil {
		return nil, fmt.Errorf("PatchDockerId: error updating docker image: %w", err)
	}

	return generated.PatchDockerId200JSONResponse{
		Success: true,
		Image: generated.DockerImage{
			DockerImage:                   image.DockerImage,
			EntropyThreshold:              float32(image.EntropyThreshold.Float64),
			LogisticGrowthRate:            float32(image.LogisticGrowthRate.Float64),
			Id:                            int(image.ID),
			MinProbability:                float32(image.MinProbability.Float64),
			Password:                      image.Password.String,
			ProbabilityDecreaseMultiplier: float32(image.ProbabilityDecreaseMultiplier.Float64),
			ProbabilityIncreaseMultiplier: float32(image.ProbabilityIncreaseMultiplier.Float64),
			ProjectId:                     int(image.ProjectID),
			Username:                      image.Username.String,
		},
	}, nil
}

func (server *serverHandler) PostDocker(ctx context.Context, request generated.PostDockerRequestObject) (generated.PostDockerResponseObject, error) {
	_, _, response, err := checkUserHasProjectPermission[generated.PostDocker401JSONResponse](server, ctx, int64(request.Body.ProjectId), authorization.Admin)
	if err != nil {
		return nil, err
	}
	if response.Success == false {
		return response, nil
	}

	dockerImage, err := server.DatabaseProvider.CreateDockerImage(ctx, queries.CreateDockerImageParams{
		DockerImage: request.Body.DockerImage,
		Password:    sql.NullString{String: request.Body.Password, Valid: true},
		ProjectID:   int64(request.Body.ProjectId),
	})
	if err != nil {
		return nil, fmt.Errorf("PostDocker: error creating docker image: %w", err)
	}

	return generated.PostDocker201JSONResponse{
		Success: true,
		Image: generated.DockerImage{
			DockerImage:                   dockerImage.DockerImage,
			EntropyThreshold:              float32(dockerImage.EntropyThreshold.Float64),
			LogisticGrowthRate:            float32(dockerImage.LogisticGrowthRate.Float64),
			MinProbability:                float32(dockerImage.MinProbability.Float64),
			Password:                      dockerImage.Password.String,
			ProbabilityDecreaseMultiplier: float32(dockerImage.ProbabilityDecreaseMultiplier.Float64),
			ProbabilityIncreaseMultiplier: float32(dockerImage.ProbabilityIncreaseMultiplier.Float64),
			Id:                            int(dockerImage.ID),
		},
	}, nil
}
