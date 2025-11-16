package core

import "time"

type SyncState struct {
	ID           string
	LastSyncTime time.Time
}
