// Package data tool for db
package data

import (
	"io"

	"github.com/DimKa163/keeper/internal/datatool"
)

type FileProvider struct {
	fp *datatool.FileProvider
}

func NewFileProvider(fp *datatool.FileProvider) *FileProvider {
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
