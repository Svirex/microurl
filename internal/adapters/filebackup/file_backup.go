package filebackup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
)

// FileBackupReader - структура reader.
type FileBackupReader struct {
	reader *json.Decoder
}

// NewFileBackupReader - создание нового reader.
func NewFileBackupReader(file *os.File) *FileBackupReader {
	return &FileBackupReader{
		reader: json.NewDecoder(file),
	}
}

var _ ports.BackupReader = (*FileBackupReader)(nil)

// Next - есть ли следующая запись.
func (reader *FileBackupReader) Next() bool {
	return reader.reader.More()
}

// Read - прочитать следующую запись
func (reader *FileBackupReader) Read(ctx context.Context) (*domain.BackupRecord, error) {
	if reader.Next() {
		record := &domain.BackupRecord{}
		err := reader.reader.Decode(record)
		if err != nil {
			return nil, fmt.Errorf("file backup reader, read: %w", err)
		}
		return record, nil
	}
	return nil, io.EOF
}

// Restore - восстановить данные из файла
func (reader *FileBackupReader) Restore(ctx context.Context, repo ports.ShortenerRepository) error {
	record, err := reader.Read(ctx)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("restore data, read: %w", err)
	}
	for record != nil {
		repo.Add(context.Background(), record.ShortID, &domain.Record{
			UID: record.UID,
			URL: record.URL,
		})
		record, err = reader.Read(ctx)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("restore data, while read: %w", err)
		}
	}
	return nil
}

// FileBackupWriter - writer в файл
type FileBackupWriter struct {
	writer *json.Encoder
}

var _ ports.BackupWriter = (*FileBackupWriter)(nil)

// NewFileBackupWriter - новый writer
func NewFileBackupWriter(file *os.File) *FileBackupWriter {
	return &FileBackupWriter{
		writer: json.NewEncoder(file),
	}
}

// Write - записать в файл
func (writer *FileBackupWriter) Write(ctx context.Context, record *domain.BackupRecord) error {
	return writer.writer.Encode(record)
}
