package file

import (
	"context"
	"time"

	"github.com/google/uuid"
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
		Event:   "upload",
		UserID:  "1",
		FileURL: file.URL,
	})
	err = uc.repo.PublishMessage(ctx, file)
	if err != nil {
		return err
	}
	time.Sleep(time.Second * 2)
	uc.notifyUsecase.Notify(notifier.Notify{
		Event:   "wait processing",
		UserID:  "1",
		FileURL: file.URL,
	})
	return err
}

func (uc *Usecase) GetFiles(ctx context.Context, query string) ([]file.File, error) {
	return uc.repo.GetFiles(ctx, query)
}

func (uc *Usecase) CompleteProcessingFile(ctx context.Context, uuidFile uuid.UUID) error {
	file, err := uc.repo.GetFileByID(ctx, uuidFile)
	if err != nil {
		return err
	}

	file.Status = "processed"

	err = uc.repo.Update(ctx, file)
	if err != nil {
		return err
	}

	uc.notifyUsecase.Notify(notifier.Notify{
		Event: "complete-processing",
		// UserID: string(file.UserID),
		UserID:  "1",
		FileURL: file.URL,
	})
	return nil
}
