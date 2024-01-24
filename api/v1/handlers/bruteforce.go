package handlers

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

func (server *serverHandler) GetProjectIdBruteforcePasswords(ctx context.Context, request generated.GetProjectIdBruteforcePasswordsRequestObject) (generated.GetProjectIdBruteforcePasswordsResponseObject, error) {
	lastid := -1
	if request.Params.LastPasswordId != nil {
		lastid = int(*request.Params.LastPasswordId)
	}

	count, err := server.DatabaseProvider.GetBruteforcePasswordsForProjectCount(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	var results []generated.BruteforcePassword
	total := bruteforcePasswordsPerPage

	if lastid < 0 {
		specificPasswords, err := server.DatabaseProvider.GetBruteforcePasswordsSpecificForProject(ctx, request.Id)
		if err != nil {
			return nil, err
		}
		total -= len(specificPasswords)

		for _, password := range specificPasswords {
			results = append(results, generated.BruteforcePassword{
				Id:       -1,
				Password: password.String,
			})
		}
	}

	if total > 0 {
		genericPasswords, err := server.DatabaseProvider.GetBruteforcePasswordsPaginated(ctx, queries.GetBruteforcePasswordsPaginatedParams{
			LastID: int64(lastid),
			Limit:  int32(total),
		})
		if err != nil {
			return nil, err
		}

		for _, password := range genericPasswords {
			results = append(results, generated.BruteforcePassword{
				Id:       int64(password.ID),
				Password: password.Password,
			})
		}
	}
	if len(results) == 0 {
		return generated.GetProjectIdBruteforcePasswords200JSONResponse{
			Success: true,
			Count:   int(count),
			Results: []generated.BruteforcePassword{},
			Next:    nil,
		}, nil
	}
	lastReturnedID := int(results[len(results)-1].Id)
	nextURL := "/api/v1/project/" + strconv.Itoa(int(request.Id)) + "/bruteforce-passwords?last_id=" + strconv.Itoa(lastReturnedID)
	return generated.GetProjectIdBruteforcePasswords200JSONResponse{
		Success: true,
		Count:   int(count),
		Next:    &nextURL,
		Results: results,
	}, nil
}

func (server *serverHandler) PatchBruteforceresultsId(ctx context.Context, request generated.PatchBruteforceresultsIdRequestObject) (generated.PatchBruteforceresultsIdResponseObject, error) {
	err := server.DatabaseProvider.UpdateScanBruteforceResult(ctx, queries.UpdateScanBruteforceResultParams{
		ID:       request.Id,
		Password: sql.NullString{String: request.Body.Password, Valid: request.Body.Password != ""},
		Tried:    int32(request.Body.Tried),
		Total:    int32(request.Body.Total),
	})
	if err == pgx.ErrNoRows {
		return generated.PatchBruteforceresultsId404JSONResponse{
			Message: "Not found",
			Success: false,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return generated.PatchBruteforceresultsId200JSONResponse{
		Success: true,
	}, nil
}

func (server *serverHandler) PostScanIdBruteforceresults(ctx context.Context, request generated.PostScanIdBruteforceresultsRequestObject) (generated.PostScanIdBruteforceresultsResponseObject, error) {
	sc, err := server.DatabaseProvider.CreateScanBruteforceResult(ctx, queries.CreateScanBruteforceResultParams{
		ScanID:   request.Id,
		Username: request.Body.Username,
		Password: sql.NullString{String: request.Body.Password, Valid: request.Body.Password != ""},
		Tried:    int32(request.Body.Tried),
		Total:    int32(request.Body.Total),
	})
	if err != nil {
		return nil, err
	}

	return generated.PostScanIdBruteforceresults200JSONResponse{
		Success: true,
		Bruteforcescanresult: &generated.BruteforceScanResult{
			Id:       int(sc.ID),
			Username: sc.Username,
			Password: sc.Password.String,
			Tried:    request.Body.Tried,
			Total:    request.Body.Total,
		},
	}, nil
}

func (server *serverHandler) GetProjectIdBruteforcedPassword(ctx context.Context, request generated.GetProjectIdBruteforcedPasswordRequestObject) (generated.GetProjectIdBruteforcedPasswordResponseObject, error) {
	pass, err := server.DatabaseProvider.GetBruteforcedPasswords(ctx, queries.GetBruteforcedPasswordsParams{
		Hash:      request.Params.Hash,
		Username:  request.Params.Username,
		ProjectID: sql.NullInt64{Int64: request.Id, Valid: true},
	})
	if err == pgx.ErrNoRows {
		return generated.GetProjectIdBruteforcedPassword404JSONResponse{
			Message: "Not found",
			Success: false,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return generated.GetProjectIdBruteforcedPassword200JSONResponse{
		Success: true,
		BruteforcedPassword: generated.BruteforcedPassword{
			Hash:             pass.Hash,
			Id:               int(pass.ID),
			LastBruteforceId: int(pass.LastBruteforceID.Int64),
			Password:         pass.Password.String,
			ProjectId:        int(pass.ProjectID.Int64),
			Username:         pass.Username,
		},
	}, nil
}

func (server *serverHandler) PatchBruteforcedPasswordsId(ctx context.Context, request generated.PatchBruteforcedPasswordsIdRequestObject) (generated.PatchBruteforcedPasswordsIdResponseObject, error) {
	_, err := server.DatabaseProvider.UpdateBruteforcedPassword(ctx, queries.UpdateBruteforcedPasswordParams{
		ID:       request.Id,
		Password: sql.NullString{String: request.Body.Password, Valid: request.Body.Password != ""},
		LastBruteforceID: sql.NullInt64{
			Int64: int64(request.Body.LastBruteforceId),
			Valid: true,
		},
	})
	if err == pgx.ErrNoRows {
		return generated.PatchBruteforcedPasswordsId404JSONResponse{
			Message: "Not found",
			Success: false,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return generated.PatchBruteforcedPasswordsId200JSONResponse{
		Success: true,
	}, nil
}

func (server *serverHandler) PostProjectIdBruteforcedPassword(ctx context.Context, request generated.PostProjectIdBruteforcedPasswordRequestObject) (generated.PostProjectIdBruteforcedPasswordResponseObject, error) {
	pass, err := server.DatabaseProvider.CreateBruteforcedPassword(ctx, queries.CreateBruteforcedPasswordParams{
		Hash:      request.Body.Hash,
		Username:  request.Body.Username,
		ProjectID: sql.NullInt64{Int64: request.Id, Valid: true},
		Password:  sql.NullString{String: request.Body.Password, Valid: request.Body.Password != ""},
		LastBruteforceID: sql.NullInt64{
			Int64: int64(request.Body.LastBruteforceId),
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return generated.PostProjectIdBruteforcedPassword200JSONResponse{
		Success: true,
		BruteforcedPassword: &generated.BruteforcedPassword{
			Hash:             request.Body.Hash,
			Id:               int(pass.ID),
			LastBruteforceId: int(pass.LastBruteforceID.Int64),
			Password:         pass.Password.String,
			ProjectId:        int(pass.ProjectID.Int64),
			Username:         pass.Username,
		},
	}, nil
}
