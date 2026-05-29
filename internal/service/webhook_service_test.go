package service

import (
	"errors"
	"testing"

	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/apperrors"
	"github.com/jalvess021/capital-pipefy/internal/domain"
	"github.com/jalvess021/capital-pipefy/internal/dto"
)

// --- mock ---

type mockEventRepo struct {
	saveIfNotExistsFn func(event *domain.ProcessedEvent) error
}

func (m *mockEventRepo) SaveIfNotExists(event *domain.ProcessedEvent) error {
	return m.saveIfNotExistsFn(event)
}

// --- mock factories ---

func newEventRepo() *mockEventRepo {
	return &mockEventRepo{
		saveIfNotExistsFn: func(event *domain.ProcessedEvent) error { return nil },
	}
}

func duplicateEventRepo() *mockEventRepo {
	return &mockEventRepo{
		saveIfNotExistsFn: func(event *domain.ProcessedEvent) error {
			return errors.Join(errors.New("duplicate"), apperrors.ErrConflict)
		},
	}
}

func clientRepoWithEmail(email string, patrimonio float64) *mockClientRepo {
	return &mockClientRepo{
		findByEmailFn: func(e string) (*domain.Client, error) {
			if e == email {
				return &domain.Client{Email: email, ValorPatrimonio: patrimonio}, nil
			}
			return nil, errors.New("not found")
		},
		saveFn:                    func(client *domain.Client) error { return nil },
		updateStatusAndPriorityFn: func(email, status, prioridade string) error { return nil },
	}
}

func absentClientRepo() *mockClientRepo {
	return &mockClientRepo{
		findByEmailFn:             func(email string) (*domain.Client, error) { return nil, errors.New("not found") },
		saveFn:                    func(client *domain.Client) error { return nil },
		updateStatusAndPriorityFn: func(email, status, prioridade string) error { return nil },
	}
}

// --- payload factory ---

func webhookRequest() dto.CardUpdatedWebhookRequest {
	return dto.CardUpdatedWebhookRequest{
		EventID:      "evt-001",
		CardID:       "card-123",
		ClienteEmail: "joao@example.com",
		Timestamp:    "2026-05-28T00:00:00Z",
	}
}

// --- tests ---

func TestProcessCardUpdated_Success(t *testing.T) {
	svc := NewWebhookService(
		clientRepoWithEmail("joao@example.com", 100_000),
		newEventRepo(),
		okPipefy(),
		zap.NewNop(),
	)

	if err := svc.ProcessCardUpdated(webhookRequest()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProcessCardUpdated_CallsUpdateStatus(t *testing.T) {
	updateCalled := false
	repo := clientRepoWithEmail("joao@example.com", 100_000)
	repo.updateStatusAndPriorityFn = func(email, status, prioridade string) error {
		updateCalled = true
		if status != "Processado" {
			t.Errorf("expected status Processado, got %s", status)
		}
		return nil
	}

	svc := NewWebhookService(repo, newEventRepo(), okPipefy(), zap.NewNop())
	svc.ProcessCardUpdated(webhookRequest())

	if !updateCalled {
		t.Error("expected UpdateStatusAndPriority to be called")
	}
}

func TestProcessCardUpdated_DuplicateEventID_ReturnsConflict(t *testing.T) {
	svc := NewWebhookService(clientRepoWithEmail("joao@example.com", 100_000), duplicateEventRepo(), okPipefy(), zap.NewNop())

	err := svc.ProcessCardUpdated(webhookRequest())

	if !errors.Is(err, apperrors.ErrConflict) {
		t.Errorf("expected ErrConflict for duplicate event_id, got %v", err)
	}
}

func TestProcessCardUpdated_ClientNotFound_ReturnsNotFound(t *testing.T) {
	svc := NewWebhookService(absentClientRepo(), newEventRepo(), okPipefy(), zap.NewNop())

	err := svc.ProcessCardUpdated(webhookRequest())

	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
