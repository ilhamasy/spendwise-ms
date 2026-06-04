package dto

import "time"

type SyncChange struct {
	EntityType string      `json:"entityType"` // transaction, category, goal, budget
	EntityID   string      `json:"entityId"`
	Operation  string      `json:"operation"` // CREATE, UPDATE, DELETE
	Payload    interface{} `json:"payload"`
	Timestamp  string      `json:"timestamp"`
}

type DataSyncRequest struct {
	LastSyncTimestamp string       `json:"lastSyncTimestamp"`
	Changes           []SyncChange `json:"changes"`
}

type DataSyncResponse struct {
	ServerChanges   []SyncChangeItem `json:"serverChanges"`
	NewSyncTimestamp string           `json:"newSyncTimestamp"`
	Conflicts       []SyncConflict   `json:"conflicts"`
}

type SyncChangeItem struct {
	EntityType string      `json:"entityType"`
	EntityID   string      `json:"entityId"`
	Data       interface{} `json:"data"`
	Timestamp  string      `json:"timestamp"`
}

type SyncConflict struct {
	EntityType   string      `json:"entityType"`
	EntityID     string      `json:"entityId"`
	LocalChange  SyncChange  `json:"localChange"`
	ServerChange interface{} `json:"serverChange"`
	Resolution   string      `json:"resolution"` // local_wins, server_wins
}

func NowTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
