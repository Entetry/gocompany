package postgre

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"entetry/gotest/internal/model"
)

// LogoRepository company logo repository interface
type LogoRepository interface {
	Create(ctx context.Context, companyID uuid.UUID, url string) error
	GetByCompanyID(ctx context.Context, companyID uuid.UUID) (*model.Logo, error)
}

// Logo company logo postgres repository struct
type Logo struct {
	db *pgxpool.Pool
}

// NewLogoRepository Creates New Logo repository object
func NewLogoRepository(db *pgxpool.Pool) *Logo {
	return &Logo{
		db: db}
}

// Create creates company logo record in db
func (l *Logo) Create(ctx context.Context, companyID uuid.UUID, url string) error {
	var logo model.Logo
	logo.ID = uuid.New()
	logo.CompanyID = companyID
	logo.Image = url
	_, err := l.db.Exec(ctx, `INSERT INTO logo (id, company_id, image) VALUES ($1, $2, $3)`, logo.ID,
		logo.CompanyID, logo.Image)

	if err != nil {
		return fmt.Errorf("cannot create Logo: %v", err)
	}
	return nil
}

// GetByCompanyID gets company logo by company uuid
func (l *Logo) GetByCompanyID(ctx context.Context, companyID uuid.UUID) (*model.Logo, error) {
	var logo model.Logo
	err := l.db.QueryRow(ctx, "SELECT id, company_id, image FROM logo WHERE company_id = $1", companyID).
		Scan(&logo.ID, &logo.CompanyID, &logo.Image)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetByCompanyID failed: %v", err)
	}
	return &logo, nil
}
