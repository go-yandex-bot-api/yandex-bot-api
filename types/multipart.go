package types

import "io"

// RequestFile represents a file to be uploaded via multipart/form-data.
type RequestFile struct {
	FieldName string
	FileName  string
	FilePath  string
	Stream    io.Reader
}

// MultipartPayload is an interface for configurations that upload files.
type MultipartPayload interface {
	Payload() any
	Files() []RequestFile
}
