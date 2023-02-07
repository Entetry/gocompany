package service

import (
	"context"
	"fmt"
	"github.com/Entetry/gocompany/internal/repository"
	"io"
	"mime/multipart"
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
	companyAlreadyHasALogoErr = "company already has a logo"
	fileSaveError             = "file save error"
	imageExt                  = ".jpeg"
)

type CompanyService interface {
	GetAll(ctx context.Context) ([]*model.Company, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Company, error)
	Create(ctx context.Context, company *model.Company) (uuid.UUID, error)
	Update(ctx context.Context, company *model.Company) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddLogo(ctx context.Context, companyID string, file *multipart.FileHeader) error
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
	return c.companyRepository.GetAll(ctx)
}

// GetByID Retrieves company based on given ID
func (c *Company) GetByID(ctx context.Context, id uuid.UUID) (*model.Company, error) {
	company, err := c.cache.Read(id)
	if err != nil {
		log.Info(err)
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
func (c *Company) Create(ctx context.Context, company *model.Company) (uuid.UUID, error) {
	return c.companyRepository.Create(ctx, company)
}

// Update update company
func (c *Company) Update(ctx context.Context, company *model.Company) error {
	return c.companyRepository.Update(ctx, company)
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
func (c *Company) AddLogo(ctx context.Context, companyID string, file *multipart.FileHeader) error {
	id, err := uuid.Parse(companyID)
	if err != nil {
		return err
	}
	logo, err := c.logoRepository.GetByCompanyID(ctx, id)
	if err != nil {
		return err
	}
	if logo != nil {
		return fmt.Errorf(companyAlreadyHasALogoErr)
	}
	imageURI := c.buildFileURI(companyID)
	err = c.saveFile(imageURI, file)
	if err != nil {
		return fmt.Errorf(fileSaveError)
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
	basepath := filepath.Join(wd, "data", "company")
	_ = os.MkdirAll(basepath, os.ModePerm)
	fileURI := filepath.Join(basepath, companyID)
	return fmt.Sprintf("%s%s", fileURI, imageExt)
}

func (c *Company) saveFile(fileName string, file *multipart.FileHeader) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		if srcError := src.Close(); srcError != nil {
			log.Printf("Error closing file: %s\n", srcError)
		}
	}()

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
