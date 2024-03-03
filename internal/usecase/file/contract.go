package file

import (
	"context"

	"github.com/mmikhail2001/test-clever-search/internal/domain/file"
)

type Repository interface {
	Upload(ctx context.Context, file file.File) (file.File, error)
	CreateFile(ctx context.Context, file file.File) error
	Search(ctx context.Context, search file.SearchQuery) ([]file.File, error)
}
