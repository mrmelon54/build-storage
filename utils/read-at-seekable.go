package utils

import (
	"io"
	"io/fs"
)

type ReadAtSeekWriterFile interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	fs.File
}
