package file

import (
	"context"

	"github.com/google/uuid"
	"github.com/mmikhail2001/test-clever-search/internal/domain/file"
)

type Usecase interface {
	Upload(ctx context.Context, file file.File) error
	GetFiles(ctx context.Context, query string) ([]file.File, error)
	CompleteProcessingFile(ctx context.Context, uuid uuid.UUID) error
}
