package handler_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jalvess021/capital-pipefy/internal/domain"
	"github.com/jalvess021/capital-pipefy/internal/handler"
	"github.com/jalvess021/capital-pipefy/internal/service"
)

// --- mocks ---

type mockEventRepo struct {
	existsByEventIDFn func(eventID string) (bool, error)
	saveFn            func(event *domain.ProcessedEvent) error
}

func (m *mockEventRepo) ExistsByEventID(eventID string) (bool, error) {
	return m.existsByEventIDFn(eventID)
}
func (m *mockEventRepo) Save(event *domain.ProcessedEvent) error { return m.saveFn(event) }

// --- repo factories ---

func newEventRepo() *mockEventRepo {
	return &mockEventRepo{
		existsByEventIDFn: func(eventID string) (bool, error) { return false, nil },
		saveFn:            func(event *domain.ProcessedEvent) error { return nil },
	}
}

func duplicateEventRepo() *mockEventRepo {
	return &mockEventRepo{
		existsByEventIDFn: func(eventID string) (bool, error) { return true, nil },
		saveFn:            func(event *domain.ProcessedEvent) error { return nil },
	}
}

func clientRepoWithEmail(email string) *mockClientRepo {
	return &mockClientRepo{
		findByEmailFn: func(e string) (*domain.Client, error) {
			if e == email {
				return &domain.Client{Email: email, ValorPatrimonio: 100_000}, nil
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

// --- helpers ---

func setupWebhookRouter(clientRepo *mockClientRepo, eventRepo *mockEventRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := service.NewWebhookService(clientRepo, eventRepo)
	h := handler.NewWebhookHandler(svc)
	r.POST("/webhooks/pipefy/card-updated", h.CardUpdated)
	return r
}

func validWebhookPayload() map[string]any {
	return map[string]any{
		"event_id":      "evt-001",
		"card_id":       "card-123",
		"cliente_email": "joao@example.com",
		"timestamp":     "2026-05-28T00:00:00Z",
	}
}

// --- tests ---

func TestWebhookHandler_CardUpdated_Success(t *testing.T) {
	w := postJSON(setupWebhookRouter(clientRepoWithEmail("joao@example.com"), newEventRepo()), "/webhooks/pipefy/card-updated", validWebhookPayload())

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestWebhookHandler_CardUpdated_DuplicateEvent_Returns200(t *testing.T) {
	w := postJSON(setupWebhookRouter(clientRepoWithEmail("joao@example.com"), duplicateEventRepo()), "/webhooks/pipefy/card-updated", validWebhookPayload())

	if w.Code != http.StatusOK {
		t.Errorf("duplicate event must return 200 (idempotent), got %d", w.Code)
	}
}

func TestWebhookHandler_CardUpdated_ClientNotFound_Returns404(t *testing.T) {
	w := postJSON(setupWebhookRouter(absentClientRepo(), newEventRepo()), "/webhooks/pipefy/card-updated", validWebhookPayload())

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestWebhookHandler_CardUpdated_InternalError_Returns500(t *testing.T) {
	eventRepo := &mockEventRepo{
		existsByEventIDFn: func(eventID string) (bool, error) { return false, errors.New("db error") },
		saveFn:            func(event *domain.ProcessedEvent) error { return nil },
	}
	w := postJSON(setupWebhookRouter(clientRepoWithEmail("joao@example.com"), eventRepo), "/webhooks/pipefy/card-updated", validWebhookPayload())

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestWebhookHandler_CardUpdated_InvalidPayload_Returns400(t *testing.T) {
	r := setupWebhookRouter(clientRepoWithEmail("joao@example.com"), newEventRepo())

	cases := []struct {
		name    string
		payload map[string]any
	}{
		{"missing event_id", func() map[string]any { p := validWebhookPayload(); delete(p, "event_id"); return p }()},
		{"missing card_id", func() map[string]any { p := validWebhookPayload(); delete(p, "card_id"); return p }()},
		{"invalid email", func() map[string]any { p := validWebhookPayload(); p["cliente_email"] = "not-an-email"; return p }()},
		{"missing timestamp", func() map[string]any { p := validWebhookPayload(); delete(p, "timestamp"); return p }()},
	}

	for _, tc := range cases {
		w := postJSON(r, "/webhooks/pipefy/card-updated", tc.payload)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", tc.name, w.Code)
		}
	}
}
