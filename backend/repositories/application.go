package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/models"
)

type ApplicationRepository struct {
	pool *pgxpool.Pool
}

func NewApplicationRepository(pool *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{pool: pool}
}

func (r *ApplicationRepository) Create(ctx context.Context, userID, jobID string, status models.ApplicationStatus) (*models.Application, error) {
	// ... database insert implementation using r.pool ...
	return &models.Application{UserID: userID, JobID: jobID, Status: status}, nil
}

func (r *ApplicationRepository) GetByID(ctx context.Context, userID, id string) (*models.Application, error) {
	// ... database select implementation using r.pool ...
	return &models.Application{}, nil
}

func (r *ApplicationRepository) UpdateStatus(ctx context.Context, userID, id string, status models.ApplicationStatus) (*models.Application, error) {
	// ... database update implementation using r.pool ...
	return &models.Application{ID: id, UserID: userID, Status: status}, nil
}

func (r *ApplicationRepository) Delete(ctx context.Context, userID, id string) error {
	// ... database delete implementation using r.pool ...
	return nil
}

func (r *ApplicationRepository) History(ctx context.Context, userID, id string) ([]models.ApplicationStatusChange, error) {
	return nil, nil
}

func (r *ApplicationRepository) ListByUserID(ctx context.Context, userID string) ([]*models.Application, error) {
	return nil, nil
}