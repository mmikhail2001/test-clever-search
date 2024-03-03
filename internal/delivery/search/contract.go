package search

import (
	"context"

	"google.golang.org/appengine/search"
)

type UseCase interface {
	Upload(ctx context.Context, search.SearchOptions)
}
