package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	clientsuc "github.com/Ishee11/isheeCRM/appointment-service/internal/usecase/clients"
	"github.com/gin-gonic/gin"
)

type clientHistoryRepo struct{}

func (clientHistoryRepo) Create(context.Context, string, string) (int, error) {
	panic("unexpected Create call")
}
func (clientHistoryRepo) FindByPhone(context.Context, string) (int, error) {
	panic("unexpected FindByPhone call")
}
func (clientHistoryRepo) Search(context.Context, string, int) ([]clientsuc.SearchResult, error) {
	panic("unexpected Search call")
}
func (clientHistoryRepo) GetInfo(context.Context, int) (clientsuc.Info, error) {
	panic("unexpected GetInfo call")
}
func (clientHistoryRepo) GetHistory(_ context.Context, clientID int) (clientsuc.History, error) {
	return clientsuc.History{ClientID: clientID, Visits: []clientsuc.Visit{}, Payments: []clientsuc.Payment{}, Subscriptions: []clientsuc.Subscription{}}, nil
}

func TestGetClientHistoryHandlerRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/clients/nope/history", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "nope"}}
	GetClientHistoryHandler(ctx)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestGetClientHistoryHandlerReturnsStableCollections(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousService := clientsService
	clientsService = clientsuc.NewService(clientHistoryRepo{})
	t.Cleanup(func() { clientsService = previousService })
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/clients/42/history", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "42"}}
	GetClientHistoryHandler(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var response struct {
		ClientID                        int `json:"client_id"`
		Visits, Payments, Subscriptions []json.RawMessage
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ClientID != 42 || response.Visits == nil || response.Payments == nil || response.Subscriptions == nil {
		t.Fatalf("invalid history response: %s", recorder.Body.String())
	}
}
