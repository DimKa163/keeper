package datatool

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FileProvider struct {
	Path string
}

func NewFileProvider(path string) *FileProvider {
	return &FileProvider{Path: path}
}

func (fp *FileProvider) IsExist(fileName string, version int32) error {
	_, err := os.Stat(buildPath(fp.Path, fileName, version))
	return err
}

func (fp *FileProvider) OpenRead(fileName string, version int32, dst ...string) (io.ReadCloser, error) {
	fullPath := buildPath(fp.Path, fileName, version, dst...)
	return os.OpenFile(fullPath, os.O_RDONLY, 0644)
}

func (fp *FileProvider) OpenWrite(fileName string, version int32, dst ...string) (io.WriteCloser, error) {
	fullPath := buildPath(fp.Path, fileName, version, dst...)
	return os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func (fp *FileProvider) Remove(fileName string, version int32, dst ...string) error {
	return os.Remove(buildPath(fp.Path, fileName, version, dst...))
}

func (fp *FileProvider) Rename(fileName string, old, new int32) error {
	return os.Rename(buildPath(fp.Path, fileName, old), buildPath(fp.Path, fileName, new))
}

func buildPath(root, name string, version int32, dst ...string) string {
	if dst == nil {
		return filepath.Join(root, fmt.Sprintf("%s_%d", name, version))
	}
	var sb strings.Builder
	for _, d := range dst {
		if sb.Len() != 0 {
			sb.WriteString("_")
		}
		sb.WriteString(d)
	}
	return filepath.Join(root, fmt.Sprintf("%s_%d_%s", name, version, sb.String()))
}
