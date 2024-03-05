package backup

import (
	"encoding/json"
	"os"

	"github.com/Svirex/microurl/internal/pkg/backup"
)

type FileBackupReader struct {
	file   *os.File
	reader *json.Decoder
}

type FileBackupWriter struct {
	file   *os.File
	writer *json.Encoder
}

var _ backup.BackupReader = (*FileBackupReader)(nil)

var _ backup.BackupWriter = (*FileBackupWriter)(nil)

func (reader *FileBackupReader) Read() (*backup.Record, error) {
	if reader.reader.More() {
		record := &backup.Record{}
		err := reader.reader.Decode(record)
		if err != nil {
			return nil, err
		}
		return record, nil
	}
	return nil, nil
}

func (reader *FileBackupReader) Close() error {
	return reader.file.Close()
}

func (writer *FileBackupWriter) Write(record *backup.Record) error {
	return writer.writer.Encode(record)
}

func (writer *FileBackupWriter) Close() error {
	return writer.file.Close()
}

func NewFileBackupReader(filename string) (backup.BackupReader, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &FileBackupReader{
		file:   file,
		reader: json.NewDecoder(file),
	}, nil
}

func NewFileBackupWriter(filename string) (backup.BackupWriter, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileBackupWriter{
		file:   file,
		writer: json.NewEncoder(file),
	}, nil
}
