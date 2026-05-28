//go:build integration

package integration_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jalvess021/capital-pipefy/internal/database"
	"github.com/jalvess021/capital-pipefy/internal/domain"
	postgresrepo "github.com/jalvess021/capital-pipefy/internal/repository/postgres"
)

// Run with: go test -tags=integration ./test/integration/... -v
// Requires: DATABASE_URL env var pointing to a real postgres instance

func setupDB(t *testing.T) *database.PostgresDB {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set — skipping integration test")
	}
	db, err := database.NewPostgres(url)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	return db
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@example.com", prefix, time.Now().UnixNano())
}

func TestClientRepository_SaveAndFind(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	repo := postgresrepo.NewClientRepository(db.GormDB())
	email := uniqueEmail("integration")

	client := &domain.Client{
		Nome:            "Integration Test User",
		Email:           email,
		TipoSolicitacao: "investimento",
		ValorPatrimonio: 300_000,
		Status:          "Aguardando Análise",
		Prioridade:      "prioridade_alta",
	}

	if err := repo.Save(client); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if client.ID == "" {
		t.Error("expected ID to be set after Save")
	}

	found, err := repo.FindByEmail(email)
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if found.Nome != client.Nome {
		t.Errorf("expected nome %s, got %s", client.Nome, found.Nome)
	}
}

func TestClientRepository_UpdateStatusAndPriority(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	repo := postgresrepo.NewClientRepository(db.GormDB())
	email := uniqueEmail("update")

	client := &domain.Client{
		Nome:            "Update Test User",
		Email:           email,
		TipoSolicitacao: "investimento",
		ValorPatrimonio: 100_000,
		Status:          "Aguardando Análise",
		Prioridade:      "prioridade_normal",
	}
	if err := repo.Save(client); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := repo.UpdateStatusAndPriority(email, "Processado", "prioridade_alta"); err != nil {
		t.Fatalf("UpdateStatusAndPriority failed: %v", err)
	}

	updated, err := repo.FindByEmail(email)
	if err != nil {
		t.Fatalf("FindByEmail after update failed: %v", err)
	}
	if updated.Status != "Processado" {
		t.Errorf("expected status Processado, got %s", updated.Status)
	}
	if updated.Prioridade != "prioridade_alta" {
		t.Errorf("expected prioridade_alta, got %s", updated.Prioridade)
	}
}
