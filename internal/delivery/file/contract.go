package file

import (
	"context"

	"github.com/mmikhail2001/test-clever-search/internal/domain/file"
)

type Usecase interface {
	Upload(ctx context.Context, file file.File) error
	Search(ctx context.Context, search file.SearchQuery) ([]file.File, error)
}
