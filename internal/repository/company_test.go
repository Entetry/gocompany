package repository

import (
	"context"
	"github.com/Entetry/gocompany/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	id1     = uuid.New()
	company = model.Company{
		ID:   id1,
		Name: "Google",
	}
)

func TestCompany_Create(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := dbPool.Exec(ctx, "TRUNCATE table company")
		require.NoError(t, err)
	}()
	t.Log("Given the need to test create company.")
	id, err := companyRepository.Create(ctx, &company)
	require.NoError(t, err, "tested create function error")
	one, err := companyRepository.GetOne(ctx, id)
	require.NoError(t, err, "tested get function error")
	require.Equal(t, company.Name, one.Name)

}

func TestCompany_Delete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := dbPool.Exec(ctx, "TRUNCATE table company")
		require.NoError(t, err)
	}()
	t.Log("Given the need to test delete company.")
	id, err := companyRepository.Create(ctx, &company)
	require.NoError(t, err, "tested create function error")
	err = companyRepository.Delete(ctx, id)
	require.NoError(t, err, "delete function error")
	_, err = companyRepository.GetOne(ctx, id)
	require.Error(t, echo.ErrNotFound, err)
}

func TestCompany_Update(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := dbPool.Exec(ctx, "TRUNCATE table company")
		require.NoError(t, err)
	}()
	t.Log("Given the need to test update company.")
	t.Log("Given the need to test delete company.")
	id, err := companyRepository.Create(ctx, &company)
	require.NoError(t, err, "tested create function error")
	updatedCompany := model.Company{
		ID:   id,
		Name: "Amazon",
	}
	err = companyRepository.Update(ctx, &updatedCompany)
	require.NoError(t, err, "tested update function error")
	c, err := companyRepository.GetOne(ctx, updatedCompany.ID)
	require.NoError(t, err, "get function error")
	require.Equal(t, updatedCompany.Name, c.Name)
}
