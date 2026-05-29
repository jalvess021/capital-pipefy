package service

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"github.com/jalvess021/capital-pipefy/internal/apperrors"
	"github.com/jalvess021/capital-pipefy/internal/domain"
	"github.com/jalvess021/capital-pipefy/internal/dto"
)

// --- mocks ---

type mockClientRepo struct {
	findByEmailFn             func(email string) (*domain.Client, error)
	saveFn                    func(client *domain.Client) error
	updateStatusAndPriorityFn func(email, status, prioridade string) error
}

func (m *mockClientRepo) FindByEmail(email string) (*domain.Client, error) {
	return m.findByEmailFn(email)
}
func (m *mockClientRepo) Save(client *domain.Client) error {
	return m.saveFn(client)
}
func (m *mockClientRepo) UpdateStatusAndPriority(email, status, prioridade string) error {
	return m.updateStatusAndPriorityFn(email, status, prioridade)
}

type mockPipefy struct {
	createCardFn      func(ctx context.Context, name, email string, assetValue float64) error
	updateCardFieldFn func(ctx context.Context, cardID, fieldID, value string) error
}

func (m *mockPipefy) CreateCard(ctx context.Context, name, email string, assetValue float64) error {
	return m.createCardFn(ctx, name, email, assetValue)
}
func (m *mockPipefy) UpdateCardField(ctx context.Context, cardID, fieldID, value string) error {
	return m.updateCardFieldFn(ctx, cardID, fieldID, value)
}

// --- mock factories ---

func notFoundRepo() *mockClientRepo {
	return &mockClientRepo{
		findByEmailFn:             func(email string) (*domain.Client, error) { return nil, errors.New("not found") },
		saveFn:                    func(client *domain.Client) error { return nil },
		updateStatusAndPriorityFn: func(email, status, prioridade string) error { return nil },
	}
}

func existingEmailRepo() *mockClientRepo {
	return &mockClientRepo{
		findByEmailFn:             func(email string) (*domain.Client, error) { return &domain.Client{Email: email}, nil },
		saveFn:                    func(client *domain.Client) error { return nil },
		updateStatusAndPriorityFn: func(email, status, prioridade string) error { return nil },
	}
}

func failingSaveRepo() *mockClientRepo {
	return &mockClientRepo{
		findByEmailFn:             func(email string) (*domain.Client, error) { return nil, errors.New("not found") },
		saveFn:                    func(client *domain.Client) error { return errors.New("db error") },
		updateStatusAndPriorityFn: func(email, status, prioridade string) error { return nil },
	}
}

func okPipefy() *mockPipefy {
	return &mockPipefy{
		createCardFn:      func(ctx context.Context, name, email string, assetValue float64) error { return nil },
		updateCardFieldFn: func(ctx context.Context, cardID, fieldID, value string) error { return nil },
	}
}

func failingPipefy() *mockPipefy {
	return &mockPipefy{
		createCardFn:      func(ctx context.Context, name, email string, assetValue float64) error { return errors.New("pipefy error") },
		updateCardFieldFn: func(ctx context.Context, cardID, fieldID, value string) error { return nil },
	}
}

// --- payload factory ---

func validRequest() dto.CreateClientRequest {
	return dto.CreateClientRequest{
		Nome:            "João Silva",
		Email:           "joao@example.com",
		TipoSolicitacao: "investimento",
		ValorPatrimonio: 100_000,
	}
}

// --- tests ---

func TestCreateClient_Success(t *testing.T) {
	svc := NewClientService(notFoundRepo(), okPipefy(), zap.NewNop())

	resp, err := svc.Create(validRequest())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Email != "joao@example.com" {
		t.Errorf("expected email joao@example.com, got %s", resp.Email)
	}
	if resp.Status != "Aguardando Análise" {
		t.Errorf("expected status 'Aguardando Análise', got %s", resp.Status)
	}
}

func TestCreateClient_CallsSave(t *testing.T) {
	saveCalled := false
	repo := notFoundRepo()
	repo.saveFn = func(client *domain.Client) error {
		saveCalled = true
		if client.Email != "joao@example.com" {
			t.Errorf("Save called with wrong email: %s", client.Email)
		}
		return nil
	}

	svc := NewClientService(repo, okPipefy(), zap.NewNop())
	svc.Create(validRequest())

	if !saveCalled {
		t.Error("expected repo.Save to be called")
	}
}

func TestCreateClient_DuplicateEmail_ReturnsConflict(t *testing.T) {
	svc := NewClientService(existingEmailRepo(), okPipefy(), zap.NewNop())

	_, err := svc.Create(validRequest())

	if !errors.Is(err, apperrors.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestCreateClient_SaveFails_ReturnsInternal(t *testing.T) {
	svc := NewClientService(failingSaveRepo(), okPipefy(), zap.NewNop())

	_, err := svc.Create(validRequest())

	if !errors.Is(err, apperrors.ErrInternal) {
		t.Errorf("expected ErrInternal on save failure, got %v", err)
	}
}

func TestCreateClient_PipefyFails_StillReturnsClient(t *testing.T) {
	svc := NewClientService(notFoundRepo(), failingPipefy(), zap.NewNop())

	resp, err := svc.Create(validRequest())

	if err != nil {
		t.Fatalf("expected no error (best-effort), got %v", err)
	}
	if resp == nil {
		t.Fatal("expected client response, got nil")
	}
}

// --- priority rule ---

func TestCalculatePriority(t *testing.T) {
	cases := []struct {
		value    float64
		expected string
	}{
		{0, "prioridade_normal"},
		{199_999, "prioridade_normal"},
		{199_999.99, "prioridade_normal"},
		{200_000, "prioridade_alta"},
		{200_000.01, "prioridade_alta"},
		{999_999, "prioridade_alta"},
	}

	for _, tc := range cases {
		got := calculatePriority(tc.value)
		if got != tc.expected {
			t.Errorf("calculatePriority(%.2f) = %s, want %s", tc.value, got, tc.expected)
		}
	}
}

func TestCreateClient_PriorityAppliedOnCreate(t *testing.T) {
	cases := []struct {
		value    float64
		expected string
	}{
		{100_000, "prioridade_normal"},
		{200_000, "prioridade_alta"},
	}

	for _, tc := range cases {
		req := validRequest()
		req.ValorPatrimonio = tc.value

		svc := NewClientService(notFoundRepo(), okPipefy(), zap.NewNop())
		resp, err := svc.Create(req)
		if err != nil {
			t.Fatalf("valor %.0f: unexpected error %v", tc.value, err)
		}
		if resp.Prioridade != tc.expected {
			t.Errorf("valor %.0f: expected %s, got %s", tc.value, tc.expected, resp.Prioridade)
		}
	}
}
