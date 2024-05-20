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
			DockerImage: dockerImage.DockerImage,
			Id:          int(dockerImage.ID),
		}
		if dockerImage.Username.Valid {
			dockerImagesResponse[i].Username = &dockerImage.Username.String
		}
		if dockerImage.Password.Valid {
			dockerImagesResponse[i].Password = &dockerImage.Password.String
		}
		if dockerImage.MinProbability.Valid {
			value := float32(dockerImage.MinProbability.Float64)
			dockerImagesResponse[i].MinProbability = &value
		}
		if dockerImage.ProbabilityDecreaseMultiplier.Valid {
			value := float32(dockerImage.ProbabilityDecreaseMultiplier.Float64)
			dockerImagesResponse[i].ProbabilityDecreaseMultiplier = &value
		}
		if dockerImage.ProbabilityIncreaseMultiplier.Valid {
			value := float32(dockerImage.ProbabilityIncreaseMultiplier.Float64)
			dockerImagesResponse[i].ProbabilityIncreaseMultiplier = &value
		}
		if dockerImage.EntropyThreshold.Valid {
			value := float32(dockerImage.EntropyThreshold.Float64)
			dockerImagesResponse[i].EntropyThreshold = &value
		}
		if dockerImage.LogisticGrowthRate.Valid {
			value := float32(dockerImage.LogisticGrowthRate.Float64)
			dockerImagesResponse[i].LogisticGrowthRate = &value
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

	dbLayers, err := server.DatabaseProvider.GetDockerLayersAndResultsForImage(ctx, dockerImage.ID)
	if err != nil {
		return nil, fmt.Errorf("GetDockerId: error getting docker layer scans: %w", err)
	}

	layers := []generated.DockerLayer{}
	layerResults := map[int64][]generated.DockerLayerResult{}

	for _, dbLayer := range dbLayers {
		if _, ok := layerResults[dbLayer.LayerID.Int64]; !ok {
			layerResults[dbLayer.LayerID.Int64] = []generated.DockerLayerResult{}
			layers = append(layers, generated.DockerLayer{
				Id:        int(dbLayer.LayerID.Int64),
				ImageId:   int(dbLayer.ImageID),
				ScannedAt: dbLayer.ScannedAt.Time.Format(time.RFC3339Nano),
				LayerHash: dbLayer.LayerHash,
				Results:   []generated.DockerLayerResult{},
			})
		}

		layerResults[dbLayer.LayerID.Int64] = append(layerResults[dbLayer.LayerID.Int64], generated.DockerLayerResult{
			CreatedAt:     dbLayer.CreatedAt.Time.Format(time.RFC3339),
			Filename:      dbLayer.Filename.String,
			Id:            int(dbLayer.ID.Int64),
			Layer:         int(dbLayer.LayerID.Int64),
			Line:          dbLayer.Line.String,
			LineNumber:    int(dbLayer.LineNumber.Int32),
			Match:         dbLayer.Match.String,
			Name:          dbLayer.Name.String,
			Password:      dbLayer.Password.String,
			Probability:   float32(dbLayer.Probability.Float64),
			ProjectId:     int(dbLayer.ProjectID.Int64),
			Username:      dbLayer.Username.String,
			PreviousLines: dbLayer.PreviousLines.String,
		})
	}

	for i, layer := range layers {
		layers[i].Results = layerResults[int64(layer.Id)]
	}

	image := generated.DockerImage{
		DockerImage: dockerImage.DockerImage,
		Id:          int(dockerImage.ID),
	}

	if dockerImage.Username.Valid {
		image.Username = &dockerImage.Username.String
	}
	if dockerImage.Password.Valid {
		image.Password = &dockerImage.Password.String
	}
	if dockerImage.MinProbability.Valid {
		value := float32(dockerImage.MinProbability.Float64)
		image.MinProbability = &value
	}
	if dockerImage.ProbabilityDecreaseMultiplier.Valid {
		value := float32(dockerImage.ProbabilityDecreaseMultiplier.Float64)
		image.ProbabilityDecreaseMultiplier = &value
	}
	if dockerImage.ProbabilityIncreaseMultiplier.Valid {
		value := float32(dockerImage.ProbabilityIncreaseMultiplier.Float64)
		image.ProbabilityIncreaseMultiplier = &value
	}
	if dockerImage.EntropyThreshold.Valid {
		value := float32(dockerImage.EntropyThreshold.Float64)
		image.EntropyThreshold = &value
	}
	if dockerImage.LogisticGrowthRate.Valid {
		value := float32(dockerImage.LogisticGrowthRate.Float64)
		image.LogisticGrowthRate = &value
	}

	return generated.GetDockerId200JSONResponse{
		Success: true,
		Image:   image,
		Layers:  layers,
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

	responseImage := generated.DockerImage{
		DockerImage: image.DockerImage,
		Id:          int(image.ID),
		ProjectId:   int(image.ProjectID),
	}

	if image.Username.Valid {
		responseImage.Username = &image.Username.String
	}
	if image.Password.Valid {
		responseImage.Password = &image.Password.String
	}
	if image.MinProbability.Valid {
		value := float32(image.MinProbability.Float64)
		responseImage.MinProbability = &value
	}
	if image.ProbabilityDecreaseMultiplier.Valid {
		value := float32(image.ProbabilityDecreaseMultiplier.Float64)
		responseImage.ProbabilityDecreaseMultiplier = &value
	}
	if image.ProbabilityIncreaseMultiplier.Valid {
		value := float32(image.ProbabilityIncreaseMultiplier.Float64)
		responseImage.ProbabilityIncreaseMultiplier = &value
	}
	if image.EntropyThreshold.Valid {
		value := float32(image.EntropyThreshold.Float64)
		responseImage.EntropyThreshold = &value
	}
	if image.LogisticGrowthRate.Valid {
		value := float32(image.LogisticGrowthRate.Float64)
		responseImage.LogisticGrowthRate = &value
	}

	return generated.PatchDockerId200JSONResponse{
		Success: true,
		Image:   responseImage,
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

	params := queries.CreateDockerImageParams{
		DockerImage: request.Body.DockerImage,
		ProjectID:   int64(request.Body.ProjectId),
	}
	if request.Body.Username != nil {
		params.Username = sql.NullString{String: *request.Body.Username, Valid: true}
	}
	if request.Body.Password != nil {
		params.Password = sql.NullString{String: *request.Body.Password, Valid: true}
	}

	dockerImage, err := server.DatabaseProvider.CreateDockerImage(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("PostDocker: error creating docker image: %w", err)
	}

	responseImage := generated.DockerImage{
		DockerImage: dockerImage.DockerImage,
		Id:          int(dockerImage.ID),
	}

	if dockerImage.Username.Valid {
		responseImage.Username = &dockerImage.Username.String
	}
	if dockerImage.Password.Valid {
		responseImage.Password = &dockerImage.Password.String
	}
	if dockerImage.MinProbability.Valid {
		value := float32(dockerImage.MinProbability.Float64)
		responseImage.MinProbability = &value
	}
	if dockerImage.ProbabilityDecreaseMultiplier.Valid {
		value := float32(dockerImage.ProbabilityDecreaseMultiplier.Float64)
		responseImage.ProbabilityDecreaseMultiplier = &value
	}
	if dockerImage.ProbabilityIncreaseMultiplier.Valid {
		value := float32(dockerImage.ProbabilityIncreaseMultiplier.Float64)
		responseImage.ProbabilityIncreaseMultiplier = &value
	}
	if dockerImage.EntropyThreshold.Valid {
		value := float32(dockerImage.EntropyThreshold.Float64)
		responseImage.EntropyThreshold = &value
	}
	if dockerImage.LogisticGrowthRate.Valid {
		value := float32(dockerImage.LogisticGrowthRate.Float64)
		responseImage.LogisticGrowthRate = &value
	}

	return generated.PostDocker201JSONResponse{
		Success: true,
		Image:   responseImage,
	}, nil
}
