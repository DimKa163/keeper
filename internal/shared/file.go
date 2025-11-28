package shared

import (
	"fmt"
	"io"
	"os"
)

type FileMode uint32

const (
	Read FileMode = iota
	Write
)

type FileProvider struct {
	path string
}

func NewFileProvider(path string) *FileProvider {
	return &FileProvider{path: path}
}

func (fp *FileProvider) IsExist(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if stat.Mode().IsDir() {
		return false, nil
	}
	return true, nil
}

func (fp *FileProvider) OpenRead(fileName string) (io.ReadCloser, error) {
	fullPath := fmt.Sprintf("%s\\%s", fp.path, fileName)
	return os.OpenFile(fullPath, os.O_RDONLY, 0644)
}

func (fp *FileProvider) OpenWrite(fileName string) (io.WriteCloser, error) {
	fullPath := fmt.Sprintf("%s\\%s", fp.path, fileName)
	return os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func (fp *FileProvider) Remove(fileName string) error {
	return os.Remove(fmt.Sprintf("%s\\%s", fp.path, fileName))
}
