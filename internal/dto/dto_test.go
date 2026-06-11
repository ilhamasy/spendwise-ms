package dto

import (
	"encoding/json"
	"testing"
)

func TestRegisterRequest_JSON(t *testing.T) {
	req := RegisterRequest{Name: "Test", Email: "test@test.com", Password: "password123"}
	data, _ := json.Marshal(req)
	var parsed RegisterRequest
	json.Unmarshal(data, &parsed)
	if parsed.Email != "test@test.com" {
		t.Error("Email mismatch")
	}
}

func TestLoginRequest_JSON(t *testing.T) {
	req := LoginRequest{Email: "test@test.com", Password: "pass"}
	data, _ := json.Marshal(req)
	var parsed LoginRequest
	json.Unmarshal(data, &parsed)
	if parsed.Email != "test@test.com" {
		t.Error("Email mismatch")
	}
}

func TestAuthResponse_JSON(t *testing.T) {
	resp := AuthResponse{
		AccessToken:  "at",
		RefreshToken: "rt",
		User:         UserInfo{ID: "1", Name: "Test", Email: "test@test.com"},
	}
	data, _ := json.Marshal(resp)
	var parsed AuthResponse
	json.Unmarshal(data, &parsed)
	if parsed.AccessToken != "at" || parsed.User.Name != "Test" {
		t.Error("AuthResponse mismatch")
	}
}

func TestErrorResponse_JSON(t *testing.T) {
	resp := ErrorResponse{Error: "unauthorized", Message: "Invalid credentials"}
	data, _ := json.Marshal(resp)
	if string(data) == "" {
		t.Error("ErrorResponse should marshal")
	}
}

func TestNowTimestamp(t *testing.T) {
	ts := NowTimestamp()
	if ts == "" {
		t.Error("Timestamp should not be empty")
	}
}

func TestDataSyncRequest_JSON(t *testing.T) {
	req := DataSyncRequest{
		LastSyncTimestamp: "2026-01-01T00:00:00Z",
		Changes:           []SyncChange{},
	}
	data, _ := json.Marshal(req)
	var parsed DataSyncRequest
	json.Unmarshal(data, &parsed)
	if parsed.LastSyncTimestamp != "2026-01-01T00:00:00Z" {
		t.Error("LastSyncTimestamp mismatch")
	}
}

func TestDataSyncResponse_JSON(t *testing.T) {
	resp := DataSyncResponse{
		ServerChanges:   []SyncChangeItem{},
		NewSyncTimestamp: "2026-01-01T00:00:00Z",
		Conflicts:       []SyncConflict{},
	}
	data, _ := json.Marshal(resp)
	var parsed DataSyncResponse
	json.Unmarshal(data, &parsed)
	if parsed.NewSyncTimestamp != "2026-01-01T00:00:00Z" {
		t.Error("NewSyncTimestamp mismatch")
	}
}
