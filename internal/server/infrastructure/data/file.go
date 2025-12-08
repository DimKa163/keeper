package data

import (
	"github.com/DimKa163/keeper/internal/shared"
	"io"
)

type FileProvider struct {
	fp *shared.FileProvider
}

func NewFileProvider(fp *shared.FileProvider) *FileProvider {
	return &FileProvider{fp: fp}
}

func (f *FileProvider) OpenRead(fileName string, version int32, dst ...string) (io.ReadCloser, error) {
	return f.fp.OpenRead(fileName, version, dst...)
}

func (f *FileProvider) OpenWrite(fileName string, version int32, dst ...string) (io.WriteCloser, error) {
	return f.fp.OpenWrite(fileName, version, dst...)
}

func (f *FileProvider) Remove(fileName string, version int32, dst ...string) error {
	return f.fp.Remove(fileName, version, dst...)
}
