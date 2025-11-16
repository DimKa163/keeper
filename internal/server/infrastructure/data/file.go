package data

import (
	"github.com/DimKa163/keeper/internal/server/domain"
)

type FileProvider struct {
	uow  domain.UnitOfWork
	path string
}
