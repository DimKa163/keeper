package cli

import (
	"database/sql"
	"github.com/DimKa163/keeper/internal/cli/core"
	"github.com/DimKa163/keeper/internal/cli/persistence"
)

type ServiceContainer struct {
	DB               *sql.DB
	RecordRepository *persistence.RecordRepository
	UserRepository   *persistence.UserRepository
	UserService      *UserService
	SyncService      *SyncService
	DataService      *DataService
	Decoder          core.Decoder
	Encoder          core.Encoder
	Console          *Console
}
