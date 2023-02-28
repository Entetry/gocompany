package handlers

import (
	"bytes"
	"context"
	"github.com/Entetry/gocompany/internal/model"
	"github.com/Entetry/gocompany/protocol/companyService"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"github.com/Entetry/gocompany/internal/service"
)

// Company handler company struct
type Company struct {
	companyService.UnimplementedCompanyServiceServer
	companyService service.CompanyService
}

// NewCompany creates new company handler
func NewCompany(companyService *service.Company) *Company {
	return &Company{companyService: companyService}
}

// GetAll Retrieves all companies
func (c *Company) GetAll(ctx context.Context, request *companyService.GetAllRequest) (*companyService.GetAllResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "input arguments are invalid")
	}

	companies, err := c.companyService.GetAll(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toGetAllResponse(companies), nil
}

func toGetAllResponse(companies []*model.Company) *companyService.GetAllResponse {
	cmps := make([]*companyService.GetCompanyResponse, 0, len(companies))
	for _, c := range companies {
		cmps = append(cmps, &companyService.GetCompanyResponse{
			Uuid: c.ID.String(),
			Name: c.Name,
		})
	}
	return &companyService.GetAllResponse{
		Companies: cmps,
	}
}

// GetByID Retrieves company based on given ID
func (c *Company) GetByID(ctx context.Context, request *companyService.GetCompanyRequest) (*companyService.GetCompanyResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "input arguments are invalid")
	}

	id, err := uuid.Parse(request.Uuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	company, err := c.companyService.GetByID(ctx, id)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if company == nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &companyService.GetCompanyResponse{
		Uuid: company.ID.String(),
		Name: company.Name,
	}, nil
}

// Create create company
func (c *Company) Create(ctx context.Context, request *companyService.CreateCompanyRequest) (*companyService.CreateCompanyResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "input arguments are invalid")
	}

	id, err := c.companyService.Create(ctx, request.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &companyService.CreateCompanyResponse{
		Uuid: id.String(),
	}, nil
}

// Update godoc company
func (c *Company) Update(ctx context.Context, request *companyService.UpdateCompanyRequest) (*companyService.UpdateCompanyResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "input arguments are invalid")
	}
	id, err := uuid.Parse(request.Uuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = c.companyService.Update(ctx, id, request.Name); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &companyService.UpdateCompanyResponse{}, nil
}

// Delete company based on given ID
func (c *Company) Delete(ctx context.Context, request *companyService.DeleteCompanyRequest) (*companyService.DeleteCompanyResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "input arguments are invalid")
	}

	id, err := uuid.Parse(request.Uuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = c.companyService.Delete(ctx, id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &companyService.DeleteCompanyResponse{}, nil
}

// GetLogoByCompanyID Retrieves company logo based on given company ID
func (c *Company) GetLogoByCompanyID(ctx echo.Context) error {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	logo, err := c.companyService.GetLogo(ctx.Request().Context(), id)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if logo == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	return ctx.File(logo)
}

// AddLogo add new company logo
func (c *Company) AddLogo(stream companyService.CompanyService_AddLogoServer) error {
	var buff bytes.Buffer
	req, err := stream.Recv()
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	companyId := req.GetCompanyID()
	for {
		req, err = stream.Recv()
		if err == io.EOF {
			if err = c.companyService.AddLogo(stream.Context(), companyId, buff.Bytes()); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			return stream.SendAndClose(&companyService.AddCompanyLogoResponse{})
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		if _, err := buff.Write(req.GetImageChunk()); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
}
