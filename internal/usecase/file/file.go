package file

import (
	"context"
	"time"

	"github.com/mmikhail2001/test-clever-search/internal/domain/file"
	"github.com/mmikhail2001/test-clever-search/internal/domain/notifier"
)

type Usecase struct {
	repo          Repository
	notifyUsecase NotifyUsecase
}

func NewUsecase(repo Repository, notifyUsecase NotifyUsecase) *Usecase {
	return &Usecase{
		repo:          repo,
		notifyUsecase: notifyUsecase,
	}
}

func (uc *Usecase) Upload(ctx context.Context, file file.File) error {
	file, err := uc.repo.Upload(ctx, file)
	if err != nil {
		return err
	}
	err = uc.repo.CreateFile(ctx, file)
	if err != nil {
		return err
	}
	uc.notifyUsecase.Notify(notifier.Notify{
		Event:  "upload",
		UserID: "1",
		Data: map[string]string{
			"url": file.URL,
		},
	})
	err = uc.repo.PublishMessage(ctx, file)
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 2)
	uc.notifyUsecase.Notify(notifier.Notify{
		Event:  "wait processing",
		UserID: "1",
		Data: map[string]string{
			"url": file.URL,
		},
	})
	return err
}

func (uc *Usecase) GetFiles(ctx context.Context, query string) ([]file.File, error) {
	return uc.repo.GetFiles(ctx, query)
}
