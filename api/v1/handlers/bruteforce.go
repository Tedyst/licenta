package handlers

import (
	"context"
	"strconv"

	"github.com/tedyst/licenta/api/v1/generated"
	"github.com/tedyst/licenta/db/queries"
)

func (server *serverHandler) GetProjectProjectidBruteforcePasswords(ctx context.Context, request generated.GetProjectProjectidBruteforcePasswordsRequestObject) (generated.GetProjectProjectidBruteforcePasswordsResponseObject, error) {
	lastid := -1
	if request.Params.LastId != nil {
		lastid = int(*request.Params.LastId)
	}

	count, err := server.DatabaseProvider.GetBruteforcePasswordsForProjectCount(ctx, request.Projectid)
	if err != nil {
		return nil, err
	}

	var results []generated.BruteforcePassword
	total := bruteforcePasswordsPerPage

	if lastid < 0 {
		specificPasswords, err := server.DatabaseProvider.GetBruteforcePasswordsSpecificForProject(ctx, request.Projectid)
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
		return generated.GetProjectProjectidBruteforcePasswords200JSONResponse{
			Success: true,
			Count:   int(count),
			Results: []generated.BruteforcePassword{},
			Next:    nil,
		}, nil
	}
	lastReturnedID := int(results[len(results)-1].Id)
	nextURL := "/api/v1/project/" + strconv.Itoa(int(request.Projectid)) + "/bruteforce-passwords?last_id=" + strconv.Itoa(lastReturnedID)
	return generated.GetProjectProjectidBruteforcePasswords200JSONResponse{
		Success: true,
		Count:   int(count),
		Next:    &nextURL,
		Results: results,
	}, nil
}
