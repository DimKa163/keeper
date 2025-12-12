// Package domain
package domain

import "io"

type Filer interface {
	OpenRead(fileName string, version int32, dst ...string) (io.ReadCloser, error)
	OpenWrite(fileName string, version int32, dst ...string) (io.WriteCloser, error)

	Remove(fileName string, version int32, dst ...string) error
}
