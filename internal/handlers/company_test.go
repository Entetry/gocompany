package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Entetry/gocompany/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	addCompany = addCompanyRequest{Name: "Google"}
)

func TestCompany_Create(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := dbPool.Exec(ctx, "TRUNCATE table company")
		require.NoError(t, err)
	}()
	t.Log("Given the need to test create company.")
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(addCompany)
	require.NoError(t, err, "failed to marhall go struct")
	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/company")
	err = companyHandler.Create(c)

	require.NoError(t, err, "Cannot create company")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestCompany_GetByID(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := dbPool.Exec(ctx, "TRUNCATE table company")
		require.NoError(t, err)
	}()
	t.Log("Given the need to test create company.")
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(addCompany)
	require.NoError(t, err, "failed to marhall go struct")

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/company")
	err = companyHandler.Create(c)
	require.NoError(t, err, "Cannot create company")
	require.Equal(t, http.StatusOK, rec.Code)
	var id uuid.UUID
	err = json.Unmarshal(rec.Body.Bytes(), &id)
	require.NoError(t, err, "Cannot unmarhsal id")
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetPath("/api/company/:id")
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	err = companyHandler.GetByID(c)
	require.NoError(t, err, "Cannot get company")
	require.Equal(t, http.StatusOK, rec.Code)
	var company model.Company
	err = json.Unmarshal(rec.Body.Bytes(), &company)
	require.NoError(t, err, "Cannot get company")
	require.Equal(t, company.Name, addCompany.Name)
}
