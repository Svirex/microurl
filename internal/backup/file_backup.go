package backup

import (
	"encoding/json"
	"os"

	"github.com/Svirex/microurl/internal/pkg/backup"
)

type FileBackup struct {
	needBackup bool
	file       *os.File
	reader     *json.Decoder
	writer     *json.Encoder
}

var _ backup.Backup = (*FileBackup)(nil)

func (fb *FileBackup) Write(record *backup.Record) error {
	if !fb.needBackup {
		return nil
	}
	err := fb.writer.Encode(record)
	if err != nil {
		return err
	}
	return nil
}

func (fb *FileBackup) Read() (*backup.Record, error) {
	if !fb.needBackup {
		return nil, nil
	}
	if fb.reader.More() {
		record := &backup.Record{}
		err := fb.reader.Decode(record)
		if err != nil {
			return nil, err
		}
		return record, nil
	}
	return nil, nil
}
