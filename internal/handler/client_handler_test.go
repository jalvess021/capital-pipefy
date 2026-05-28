package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jalvess021/capital-pipefy/internal/apperrors"
	"github.com/jalvess021/capital-pipefy/internal/domain"
	"github.com/jalvess021/capital-pipefy/internal/dto"
	"github.com/jalvess021/capital-pipefy/internal/handler"
	"github.com/jalvess021/capital-pipefy/internal/service"
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
func (m *mockClientRepo) Save(client *domain.Client) error { return m.saveFn(client) }
func (m *mockClientRepo) UpdateStatusAndPriority(email, status, prioridade string) error {
	return m.updateStatusAndPriorityFn(email, status, prioridade)
}

type mockPipefy struct {
	createCardFn func(ctx context.Context, name, email string, assetValue float64) error
}

func (m *mockPipefy) CreateCard(ctx context.Context, name, email string, assetValue float64) error {
	return m.createCardFn(ctx, name, email, assetValue)
}
func (m *mockPipefy) UpdateCardField(ctx context.Context, cardID, fieldID, value string) error {
	return nil
}

// --- repo factories ---

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
	return &mockPipefy{createCardFn: func(ctx context.Context, name, email string, assetValue float64) error { return nil }}
}

func failingPipefy() *mockPipefy {
	return &mockPipefy{createCardFn: func(ctx context.Context, name, email string, assetValue float64) error {
		return fmt.Errorf("pipefy error: %w", apperrors.ErrInternal)
	}}
}

// --- helpers ---

func setupClientRouter(repo *mockClientRepo, pipefy *mockPipefy) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := service.NewClientService(repo, pipefy)
	h := handler.NewClientHandler(svc)
	r.POST("/clientes", h.Create)
	return r
}

func postJSON(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func validClientPayload() map[string]any {
	return map[string]any{
		"cliente_nome":     "João Silva",
		"cliente_email":    "joao@example.com",
		"tipo_solicitacao": "investimento",
		"valor_patrimonio": 100000,
	}
}

// --- tests ---

func TestClientHandler_Create_Success(t *testing.T) {
	w := postJSON(setupClientRouter(notFoundRepo(), okPipefy()), "/clientes", validClientPayload())

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp dto.ClientResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Email != "joao@example.com" {
		t.Errorf("expected email joao@example.com, got %s", resp.Email)
	}
}

func TestClientHandler_Create_DuplicateEmail_Returns409(t *testing.T) {
	w := postJSON(setupClientRouter(existingEmailRepo(), okPipefy()), "/clientes", validClientPayload())

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestClientHandler_Create_InternalError_Returns500(t *testing.T) {
	w := postJSON(setupClientRouter(failingSaveRepo(), okPipefy()), "/clientes", validClientPayload())

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestClientHandler_Create_InvalidPayload_Returns400(t *testing.T) {
	r := setupClientRouter(notFoundRepo(), okPipefy())

	cases := []struct {
		name    string
		payload map[string]any
	}{
		{"missing nome", func() map[string]any { p := validClientPayload(); delete(p, "cliente_nome"); return p }()},
		{"missing email", func() map[string]any { p := validClientPayload(); delete(p, "cliente_email"); return p }()},
		{"invalid email", func() map[string]any { p := validClientPayload(); p["cliente_email"] = "not-an-email"; return p }()},
		{"missing patrimonio", func() map[string]any { p := validClientPayload(); delete(p, "valor_patrimonio"); return p }()},
	}

	for _, tc := range cases {
		w := postJSON(r, "/clientes", tc.payload)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", tc.name, w.Code)
		}
	}
}
