// Package postgre contains postgre repository structs
package postgre

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	"entetry/gotest/internal/model"
)

// CompanyRepository interface for company repository
type CompanyRepository interface {
	Create(ctx context.Context, company *model.Company) (uuid.UUID, error)
	Update(ctx context.Context, company *model.Company) error
	Delete(ctx context.Context, uuid uuid.UUID) error
	GetOne(ctx context.Context, uuid uuid.UUID) (*model.Company, error)
	GetAll(ctx context.Context) ([]*model.Company, error)
}

// Company postgres company repository struct
type Company struct {
	db *pgxpool.Pool
}

// NewCompanyRepository Creates Company object
func NewCompanyRepository(db *pgxpool.Pool) *Company {
	return &Company{
		db: db,
	}
}

// GetAll get all companies from db
func (c *Company) GetAll(ctx context.Context) ([]*model.Company, error) {
	rows, err := c.db.Query(ctx, `SELECT id, name FROM company`)
	if err != nil {
		return nil, fmt.Errorf("query: %v", err)
	}
	defer rows.Close()

	var results []*model.Company

	for rows.Next() {
		var company model.Company

		err = rows.Scan(&company.ID, &company.Name)
		if err != nil {
			return nil, fmt.Errorf("scan: %v", err)
		}

		results = append(results, &company)
	}

	return results, nil
}

// GetOne gets Company by its uuid
func (c *Company) GetOne(ctx context.Context, id uuid.UUID) (*model.Company, error) {
	var company model.Company
	err := c.db.QueryRow(ctx, "SELECT id, name FROM company WHERE id = $1", id).Scan(&company.ID, &company.Name)
	return &company, err
}

// Create creates New Company record in db
func (c *Company) Create(ctx context.Context, company *model.Company) (uuid.UUID, error) {
	company.ID = uuid.New()
	_, err := c.db.Exec(ctx, "INSERT INTO company(id, name) VALUES ($1, $2) RETURNING id, name;",
		company.ID, company.Name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("cannot create Company: %v", err)
	}
	return company.ID, err
}

// Update updates company in db
func (c *Company) Update(ctx context.Context, company *model.Company) error {
	_, err := c.db.Exec(ctx, "UPDATE company SET name = $2 WHERE id=$1 RETURNING id, name;",
		company.ID, company.Name)
	if err != nil {
		return fmt.Errorf("cannot update Company: %v", err)
	}
	return err
}

// Delete deletes company from db
func (c *Company) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := c.db.Exec(ctx, "DELETE FROM company WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("cannot delete Company: %v", err)
	}
	return nil
}
