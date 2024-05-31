package ports

import (
	"context"
	"errors"

	"github.com/Svirex/microurl/internal/core/domain"
)

var ErrAlreadyExists = errors.New("already exists")
var ErrNotFound = errors.New("not found")

// ShortenerService - интерфейс сервиса сокращения ссылок.
type ShortenerService interface {
	// Add - добавить запись и вернуть сокращенный URL.
	Add(context.Context, *domain.Record) (domain.ShortURL, error)

	// Get - получить URL по сокращенному ID.
	Get(ctx context.Context, shortID domain.ShortID) (domain.URL, error)

	// Batch - добавить несколько записей.
	Batch(ctx context.Context, uid domain.UID, data []domain.BatchRecord) ([]domain.BatchRecord, error)

	// UserURLs - получить все записи для определенного пользователя
	UserURLs(ctx context.Context, uid domain.UID) ([]domain.URLData, error)

	// Shutdown - завершить сервис
	Shutdown() error
}

// DBCheckerService - интерфейс сервиса проверки соединения с БД.
type DBCheckerService interface {
	// DBCheckerService - проверка, что БД доступна.
	Ping(context.Context) error

	// Shutdown - выключение сервиса
	Shutdown() error
}

// ShortenerRepository - интерфейс репозитория сервиса сокращения ссылок.
type ShortenerRepository interface {
	// Add - добавить короткие идентификатор и URL для пользователя в БД.
	// Если такой URL уже есть, то вернуть соответствующий ему идентификатор.
	Add(ctx context.Context, shortID domain.ShortID, data *domain.Record) (domain.ShortID, error)

	// Get - получить оригинальный URL по идентификатору.
	Get(ctx context.Context, shortID domain.ShortID) (domain.URL, error)

	// Batch - записать в БД батч записей.
	Batch(ctx context.Context, uid domain.UID, data []domain.BatchRecord) ([]domain.BatchRecord, error)

	// UserURLs - вернуть все записи для пользователя
	UserURLs(ctx context.Context, uid domain.UID) ([]domain.URLData, error)

	// Shutdown - выключить сервис
	Shutdown() error
}

// BackupWriter - интерфейс для сохранения записей в файл.
type BackupWriter interface {
	Write(ctx context.Context, record *domain.BackupRecord) error
}

// BackupReader - интерфейс для чтения записей из файла.
type BackupReader interface {
	Next() bool
	Read(ctx context.Context) (*domain.BackupRecord, error)
	Restore(ctx context.Context, repo ShortenerRepository) error
}

// StringGenerator - интерфейс генератора рандомных строк.
type StringGenerator interface {
	Generate(ctx context.Context, size uint) string
}

// DeleterService - интерфейс сервиса, который помечает URL удаленными.
type DeleterService interface {
	Process(ctx context.Context, uid string, shortIDs []string)
	Run() error
	Shutdown() error
}

// DeleterRepository - репозиторий, который выполняет удаление записей в БД.
type DeleterRepository interface {
	Delete(ctx context.Context, batch []*domain.DeleteData) error
}

// DBCheck - интерфейс проверки коннекта к БД.
type DBCheck interface {
	Ping(context.Context) error
}
