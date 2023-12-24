package workerauth

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/tedyst/licenta/cache"
	"github.com/tedyst/licenta/models"
)

const workerAuthHeader = "X-Worker-Token"

type WorkerAuth interface {
	Handler(next http.Handler) http.Handler
	GetWorker(ctx context.Context) *models.Worker
}

type workerAuthQuerier interface {
	GetWorkerByToken(ctx context.Context, token string) (*models.Worker, error)
}

type workerAuthKey struct{}

type workerAuth struct {
	cache   cache.CacheProvider[models.Worker]
	querier workerAuthQuerier
}

func (wa *workerAuth) getWorkerAuthData(ctx context.Context) *models.Worker {
	if data, ok := ctx.Value(workerAuthKey{}).(*models.Worker); ok {
		return data
	}
	return nil
}

func (wa *workerAuth) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		workerAuthData := wa.getWorkerAuthData(ctx)

		if r.Header.Get(workerAuthHeader) != "" {
			worker, ok, err := wa.cache.Get(r.Header.Get(workerAuthHeader))
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !ok {
				worker, err := wa.querier.GetWorkerByToken(ctx, r.Header.Get(workerAuthHeader))
				if err != nil && err != pgx.ErrNoRows {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				if err == nil {
					wa.cache.Set(r.Header.Get(workerAuthHeader), *worker)
					workerAuthData = worker
				}
			} else {
				workerAuthData = &worker
			}
		}

		if workerAuthData != nil {
			ctx = context.WithValue(ctx, workerAuthKey{}, workerAuthData)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (wa *workerAuth) GetWorker(ctx context.Context) *models.Worker {
	return wa.getWorkerAuthData(ctx)
}

func NewWorkerAuth(cache cache.CacheProvider[models.Worker], querier workerAuthQuerier) WorkerAuth {
	return &workerAuth{
		cache:   cache,
		querier: querier,
	}
}
