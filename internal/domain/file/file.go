package file

import (
	"mime/multipart"

	"github.com/google/uuid"
)

type File struct {
	ID          uuid.UUID
	Filename    string
	Path        string
	Size        int64
	File        multipart.File
	ContentType string
	Status      string
	URL         string
}
