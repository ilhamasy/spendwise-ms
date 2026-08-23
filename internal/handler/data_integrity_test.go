package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"spendwise-ms/internal/dto"

	"github.com/gin-gonic/gin"
)

func TestDataIntegrity_SyncBatchLimitExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/transactions/sync", func(c *gin.Context) { c.Set("userId", "test-user-integrity"); c.Next() }, SyncTransactions)

	// Build a payload with 501 items to trigger the integrity batch limit
	items := make([]dto.SyncTransactionRequest, 501)
	for i := 0; i < 501; i++ {
		items[i] = dto.SyncTransactionRequest{
			Type:       "expense",
			Amount:     1000,
			CategoryID: "cat-1",
			OccurredAt: "2026-01-01",
			Note:       "Batch item",
		}
	}

	body, _ := json.Marshal(dto.SyncRequest{Transactions: items})
	req, _ := http.NewRequest("POST", "/api/transactions/sync", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request when sync batch exceeds 500 items, got %d", w.Code)
	}
}

func TestDataIntegrity_SyncNoteSanitization(t *testing.T) {
	sanitized := dto.SanitizeString("<script>alert('XSS-Data-Integrity')</script>")
	expected := "&lt;script&gt;alert(&#39;XSS-Data-Integrity&#39;)&lt;/script&gt;"
	if sanitized != expected {
		t.Errorf("Sanitization failed: expected '%s', got '%s'", expected, sanitized)
	}
}
