package types

import "io"

// RequestFile represents a file to be uploaded via multipart/form-data.
//
// ContentType is optional. If empty, the Content-Type is inferred from the
// file extension of FileName or FilePath. For generic binary files (e.g. documents)
// it falls back to "application/octet-stream".
type RequestFile struct {
	FieldName   string
	FileName    string
	FilePath    string
	Stream      io.Reader
	ContentType string // optional: explicit MIME type for the multipart part
}

// MultipartPayload is an interface for configurations that upload files.
type MultipartPayload interface {
	Payload() any
	Files() []RequestFile
}
