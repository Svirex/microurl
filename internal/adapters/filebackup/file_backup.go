package filebackup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
)

type FileBackupReader struct {
	reader *json.Decoder
}

func NewFileBackupReader(file *os.File) *FileBackupReader {
	return &FileBackupReader{
		reader: json.NewDecoder(file),
	}
}

var _ ports.BackupReader = (*FileBackupReader)(nil)

func (reader *FileBackupReader) Next() bool {
	return reader.reader.More()
}

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

type FileBackupWriter struct {
	writer *json.Encoder
}

var _ ports.BackupWriter = (*FileBackupWriter)(nil)

func NewFileBackupWriter(file *os.File) *FileBackupWriter {
	return &FileBackupWriter{
		writer: json.NewEncoder(file),
	}
}

func (writer *FileBackupWriter) Write(ctx context.Context, record *domain.BackupRecord) error {
	return writer.writer.Encode(record)
}
