package file

import (
	"context"

	"github.com/mmikhail2001/test-clever-search/internal/domain/file"
)

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

func (uc *Usecase) Upload(ctx context.Context, file file.File) error {
	file, err := uc.repo.Upload(ctx, file)
	if err != nil {
		return err
	}
	err = uc.repo.CreateFile(ctx, file)
	return err
}

func (uc *Usecase) Search(ctx context.Context, search file.SearchQuery) ([]file.File, error) {
	return uc.repo.Search(ctx, search)
}
