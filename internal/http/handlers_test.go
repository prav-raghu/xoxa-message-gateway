package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"xoxa-message-gateway/internal/config"
	"xoxa-message-gateway/internal/repo"
	"xoxa-message-gateway/internal/service"
	"xoxa-message-gateway/pkg/dto"
)

type fakePublisher struct{}

func (fakePublisher) Publish(subject string, data []byte) error { return nil }

func testServer(t *testing.T) (*gin.Engine, *config.Config) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&repo.Message{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	cfg := &config.Config{JWTSecret: "test-secret", JWTIssuer: "xoxa-gateway", JWTAudience: "internal"}
	svc := service.New(repo.NewMessageRepository(db), fakePublisher{})

	r := gin.New()
	RegisterHandlers(r, svc, cfg)
	return r, cfg
}

func validToken(t *testing.T, cfg *config.Config) string {
	t.Helper()
	claims := jwt.MapClaims{
		"iss": cfg.JWTIssuer,
		"aud": cfg.JWTAudience,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestHealthz(t *testing.T) {
	r, _ := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSendMessage_RequiresAuth(t *testing.T) {
	r, _ := testServer(t)

	body, _ := json.Marshal(dto.SendRequest{Channel: "sms", To: "+1", Text: "hi"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestSendMessage_AndGetMessage(t *testing.T) {
	r, cfg := testServer(t)
	token := validToken(t, cfg)

	body, _ := json.Marshal(dto.SendRequest{Channel: "sms", To: "+1", Text: "hi"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}

	var sendResp dto.SendResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &sendResp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if sendResp.ID == "" || sendResp.Status != "queued" {
		t.Fatalf("unexpected send response: %+v", sendResp)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/messages/"+sendResp.ID, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", getRec.Code, getRec.Body.String())
	}
}

func TestGetMessage_NotFound(t *testing.T) {
	r, cfg := testServer(t)
	token := validToken(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/msg_does_not_exist", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSendMessage_InvalidBody(t *testing.T) {
	r, cfg := testServer(t)
	token := validToken(t, cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages", bytes.NewReader([]byte(`{"channel":""}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
