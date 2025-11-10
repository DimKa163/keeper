package data

import (
	"context"

	"github.com/DimKa163/keeper/internal/server/domain"
	"github.com/beevik/guid"
)

type FileProvider struct {
	uow  domain.UnitOfWork
	path string
}

func NewFileProvider(uow domain.UnitOfWork, path string) *FileProvider {
	return &FileProvider{uow: uow, path: path}
}

func (fs *FileProvider) ExecuteWriter(ctx context.Context, dataID guid.Guid) domain.Writer {
	return &FileWriter{uow: fs.uow, dataID: dataID}
}

func (fs *FileProvider) ExecuteReader(ctx context.Context, dataID guid.Guid) domain.Reader {
	panic("implement me")
}

type FileWriter struct {
	uow    domain.UnitOfWork
	dataID guid.Guid
}

func (fw *FileWriter) Write(ctx context.Context, part *domain.FilePart) error {
	panic("implement me")
}
