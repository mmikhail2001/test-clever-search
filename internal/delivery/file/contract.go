package file

import "context"

type UseCase interface {
	Search(ctx context.Context, query string, filetype string)
}
