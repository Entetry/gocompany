package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ory/dockertest"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"testing"
)

var (
	dbPool            *pgxpool.Pool
	companyRepository *Company
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("unix:///home/entetry/.docker/desktop/docker.sock")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("postgres", "14.1-alpine", []string{"POSTGRES_PASSWORD=password123"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	var dbHostAndPort string

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pool.Retry(func() error {
		dbHostAndPort = resource.GetHostPort("5432/tcp")

		dbPool, err = pgxpool.Connect(ctx, fmt.Sprintf("postgresql://postgres:password123@%v/postgres", dbHostAndPort))
		if err != nil {
			return err
		}

		return dbPool.Ping(ctx)
	})
	if err != nil {
		cancel()
		log.Fatalf("Could not connect to database: %s", err)
	}

	companyRepository = NewCompanyRepository(dbPool)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("flyway",
		"-user=postgres",
		"-password=password123",
		"-locations=filesystem:../../migrations",
		fmt.Sprintf("-url=jdbc:postgresql://%v/postgres", dbHostAndPort),
		"migrate")

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	code := m.Run()

	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	err = resource.Expire(1)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}
