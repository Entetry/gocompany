package handlers

import (
	"context"
	"fmt"
	"github.com/Entetry/gocompany/protocol/companyService"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
)

var (
	addCompany = addCompanyRequest{Name: "Google"}
)

func TestCompany_AddLogo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := dbPool.Exec(ctx, "TRUNCATE table company")
		require.NoError(t, err)
	}()
	t.Log("Given the need to test create company.")
	id, err := companyClient.Create(ctx, &companyService.CreateCompanyRequest{Name: "Google"})
	t.Log(t, err, "Failed to create company")
	t.Log("Given the need to test add company logo")
	file, err := os.Open(buildFileURI("logo"))
	require.NoError(t, err, "Cannot open test file")
	logo, err := companyClient.AddLogo(ctx)
	require.NoError(t, err, "failed to send go struct")
	buf := make([]byte, 1024)
	err = logo.Send(&companyService.AddCompanyLogoRequest{
		Data:       &companyService.AddCompanyLogoRequest_CompanyID{CompanyID: id.Uuid},
		ImageChunk: buf[:0],
	})
	require.NoError(t, err, "failed to send go struct")
	for {
		num, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		require.NoError(t, err, "failed to send file")

		err = logo.Send(&companyService.AddCompanyLogoRequest{ImageChunk: buf[:num]})
		require.NoError(t, err, "Failed to send image chunk")
	}

	_, err = logo.CloseAndRecv()
	require.NoError(t, err, "Failed to get response")

}

func buildFileURI(name string) string {
	wd, _ := os.Getwd()
	basepath := filepath.Join(wd, "rsc")
	_ = os.MkdirAll(basepath, os.ModePerm)
	fileURI := filepath.Join(basepath, name)
	return fmt.Sprintf("%s%s", fileURI, ".jpeg")
}
