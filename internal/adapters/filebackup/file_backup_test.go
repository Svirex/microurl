package filebackup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/stretchr/testify/require"
)

func TestNextEmptyFile(t *testing.T) {
	file, err := os.CreateTemp("", "test-file-backup-")
	require.NoError(t, err)
	defer os.Remove(file.Name())
	defer file.Close()
	reader := NewFileBackupReader(file)
	require.False(t, reader.Next())
}

func TestHasNext(t *testing.T) {
	file, err := os.CreateTemp("", "test-file-backup-")
	require.NoError(t, err)
	defer os.Remove(file.Name())
	defer file.Close()
	reader := NewFileBackupReader(file)
	record := domain.BackupRecord{
		UUID:    "1",
		ShortID: "42fasfasf",
		URL:     "http://svirex.ru",
	}
	data, err := json.Marshal(&record)
	require.NoError(t, err)
	_, err = file.Write(data)
	require.NoError(t, err)
	file.Seek(0, 0)
	require.True(t, reader.Next())
}

func TestReadInvalidRecord(t *testing.T) {
	file, err := os.CreateTemp("", "test-file-backup-")
	require.NoError(t, err)
	defer os.Remove(file.Name())
	defer file.Close()
	file.Write([]byte(`{"uuid":1,"short_url":"42fasfasf","original_url":"http://svirex.ru"}`))
	file.Seek(0, 0)
	reader := NewFileBackupReader(file)
	_, err = reader.Read(context.Background())
	fmt.Println(err, errors.Is(err, &json.UnmarshalTypeError{}))
	e := &json.UnmarshalTypeError{}
	require.ErrorAs(t, err, &e)

}

func TestReadEOF(t *testing.T) {
	file, err := os.CreateTemp("", "test-file-backup-")
	require.NoError(t, err)
	defer os.Remove(file.Name())
	defer file.Close()
	reader := NewFileBackupReader(file)
	_, err = reader.Read(context.Background())
	require.ErrorIs(t, err, io.EOF)
}

func TestReadGood(t *testing.T) {
	file, err := os.CreateTemp("", "test-file-backup-")
	require.NoError(t, err)
	defer os.Remove(file.Name())
	defer file.Close()

	record := domain.BackupRecord{
		UUID:    "1",
		ShortID: "42fasfasf",
		URL:     "http://svirex.ru",
	}
	data, err := json.Marshal(&record)
	require.NoError(t, err)
	_, err = file.Write(data)
	require.NoError(t, err)

	file.Seek(0, 0)
	reader := NewFileBackupReader(file)
	r, err := reader.Read(context.Background())
	require.NoError(t, err)
	require.Equal(t, record, *r)
}

func TestWriteGood(t *testing.T) {
	file, err := os.CreateTemp("", "test-file-backup-")
	require.NoError(t, err)
	defer os.Remove(file.Name())
	defer file.Close()

	record := &domain.BackupRecord{
		UUID:    "1",
		ShortID: "42fasfasf",
		URL:     "http://svirex.ru",
	}
	writer := NewFileBackupWriter(file)
	err = writer.Write(context.Background(), record)
	require.NoError(t, err)
	file.Seek(0, 0)
	data, err := io.ReadAll(file)
	require.NoError(t, err)

	d, _ := json.Marshal(record)
	expected := []byte(string(d) + "\n")
	require.Equal(t, expected, data)
}
