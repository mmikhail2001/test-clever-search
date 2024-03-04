package file

import (
	"context"

	"github.com/google/uuid"
	"github.com/mmikhail2001/test-clever-search/internal/domain/file"
	"github.com/mmikhail2001/test-clever-search/internal/domain/notifier"
)

type Repository interface {
	Upload(ctx context.Context, file file.File) (file.File, error)
	CreateFile(ctx context.Context, file file.File) error
	GetFiles(ctx context.Context, query string) ([]file.File, error)
	PublishMessage(ctx context.Context, file file.File) error
	GetFileByID(ctx context.Context, uuidFile uuid.UUID) (file.File, error)
	Update(ctx context.Context, file file.File) error
}

type NotifyUsecase interface {
	Notify(notify notifier.Notify)
}
