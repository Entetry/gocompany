package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Entetry/gocompany/internal/repository"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/Entetry/gocompany/internal/cache"
	"github.com/Entetry/gocompany/internal/event"
	"github.com/Entetry/gocompany/internal/model"
	"github.com/Entetry/gocompany/internal/producer"
)

const (
	ErrCompanyAlreadyHasALogo = "company already has a logo"
	ErrFileSave               = "file save error"
	imageExt                  = ".jpeg"
)

type CompanyService interface {
	GetAll(ctx context.Context) ([]*model.Company, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Company, error)
	Create(ctx context.Context, name string) (uuid.UUID, error)
	Update(ctx context.Context, uuid uuid.UUID, name string) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddLogo(ctx context.Context, companyID string, file []byte) error
	GetLogo(ctx context.Context, companyID uuid.UUID) (string, error)
}

// Company service company struct
type Company struct {
	companyRepository repository.CompanyRepository
	logoRepository    repository.LogoRepository
	cache             cache.Cache
	producer          producer.Company
}

// NewCompany creates new Company service
func NewCompany(
	companyRepository repository.CompanyRepository, logoRepository repository.LogoRepository,
	localCache cache.Cache, redisProducer producer.Company) *Company {
	return &Company{
		companyRepository: companyRepository, logoRepository: logoRepository, cache: localCache, producer: redisProducer}
}

// GetAll return all companies
func (c *Company) GetAll(ctx context.Context) ([]*model.Company, error) {
	cmps, err := c.companyRepository.GetAll(ctx)
	if errors.Is(repository.ErrNotFound, err) {
		return nil, nil
	}
	return cmps, err
}

// GetByID Retrieves company based on given ID
func (c *Company) GetByID(ctx context.Context, id uuid.UUID) (*model.Company, error) {
	company, err := c.cache.Read(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, nil
	}
	if company != nil {
		return company, nil
	}
	company, err = c.companyRepository.GetOne(ctx, id)
	if company != nil {
		redisErr := c.producer.Produce(ctx, company.ID, event.UPDATE, company.Name)
		if redisErr != nil {
			log.Error(err)
		}
	}

	return company, err
}

// Create  company
func (c *Company) Create(ctx context.Context, name string) (uuid.UUID, error) {
	return c.companyRepository.Create(ctx, name)
}

// Update update company
func (c *Company) Update(ctx context.Context, uuid uuid.UUID, name string) error {
	return c.companyRepository.Update(ctx, uuid, name)
}

// Delete delete company
func (c *Company) Delete(ctx context.Context, id uuid.UUID) error {
	company, err := c.cache.Read(id)
	if err != nil {
		log.Info(err)
	}
	if company != nil {
		c.cache.Delete(id)
	}
	return c.companyRepository.Delete(ctx, id)
}

// AddLogo add logo to a company( fails if company already has a logo)
func (c *Company) AddLogo(ctx context.Context, companyID string, file []byte) error {
	id, err := uuid.Parse(companyID)

	if err != nil {
		return err
	}
	logo, err := c.logoRepository.GetByCompanyID(ctx, id)
	if err != nil {
		return err
	}
	if logo != nil {
		return fmt.Errorf(ErrCompanyAlreadyHasALogo)
	}
	imageURI := c.buildFileURI(companyID)
	err = c.saveFile(imageURI, file)
	if err != nil {
		return fmt.Errorf(ErrFileSave)
	}

	err = c.logoRepository.Create(ctx, id, imageURI)
	if err != nil {
		return err
	}
	return nil
}

// GetLogo Get company logo
func (c *Company) GetLogo(ctx context.Context, companyID uuid.UUID) (string, error) {
	logo, err := c.logoRepository.GetByCompanyID(ctx, companyID)
	if err != nil {
		return "", err
	}
	return logo.Image, nil
}

func (c *Company) buildFileURI(companyID string) string {
	wd, _ := os.Getwd()
	basepath := filepath.Join(wd, "rsc")
	_ = os.MkdirAll(basepath, os.ModePerm)
	fileURI := filepath.Join(basepath, companyID)
	return fmt.Sprintf("%s%s", fileURI, imageExt)
}

func (c *Company) saveFile(fileName string, file []byte) error {

	src := bytes.NewReader(file)

	dst, err := os.Create(filepath.Clean(fileName))
	if err != nil {
		return err
	}
	defer func() {
		if dstError := dst.Close(); dstError != nil {
			log.Printf("Error closing file: %s\n", dstError)
		}
	}()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	err = dst.Sync()
	if err != nil {
		return err
	}

	return nil
}
