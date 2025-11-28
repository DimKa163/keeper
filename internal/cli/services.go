package cli

import (
	"database/sql"
	"github.com/DimKa163/keeper/internal/cli/core"
)

type ServiceContainer struct {
	DB          *sql.DB
	UserService *UserService
	SyncService *SyncService
	DataService *DataService
	Decoder     core.Decoder
	Encoder     core.Encoder
}
