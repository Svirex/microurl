package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Svirex/microurl/internal/logging"
	"github.com/jmoiron/sqlx"
)

type DeleteData struct {
	uid     string
	shortID string
}

type DeleterService interface {
	Process(ctx context.Context, uid string, shortIDs []string)
	Run() error
	Shutdown() error
}

type DefaultDeleter struct {
	wg          sync.WaitGroup
	db          *sqlx.DB
	logger      logging.Logger
	batchSize   int
	mayShutdown chan struct{}

	errorChan chan error
	fanInChan chan *DeleteData
}

func NewDefaultDeleter(db *sqlx.DB, logger logging.Logger, batchSize int) (*DefaultDeleter, error) {
	service := &DefaultDeleter{
		db:          db,
		logger:      logger,
		batchSize:   batchSize,
		errorChan:   make(chan error, 5),
		fanInChan:   make(chan *DeleteData, batchSize),
		mayShutdown: make(chan struct{}),
	}
	return service, nil
}

var _ DeleterService = (*DefaultDeleter)(nil)

func (ds *DefaultDeleter) Run() error {
	// запустить горутину, которая пишет в базу
	// запустить горутину, которая логирует ошибки
	//
	go ds.dbWriter()
	go ds.errorLogger()

	return nil
}

func (ds *DefaultDeleter) Process(_ context.Context, uid string, shortIDs []string) {
	ds.wg.Add(1)
	go ds.generator(uid, shortIDs)
}

func (ds *DefaultDeleter) Shutdown() error {
	ds.wg.Wait()
	close(ds.fanInChan)
	<-ds.mayShutdown
	close(ds.errorChan)
	return nil
}

func (ds *DefaultDeleter) generator(uid string, shortIDs []string) {
	for _, v := range shortIDs {
		ds.fanInChan <- &DeleteData{
			uid:     uid,
			shortID: v,
		}
	}
	ds.wg.Done()
}

func (ds *DefaultDeleter) errorLogger() {
	for err := range ds.errorChan {
		ds.logger.Error("write batch err", "err", err)
	}
}

func (ds *DefaultDeleter) dbWriter() {
	batch := make([]*DeleteData, 0, ds.batchSize)
	ticker := time.NewTicker(time.Second)

	clearBatch := func(batch []*DeleteData) []*DeleteData {
		for i := range batch {
			batch[i] = nil
		}
		return batch[:0]
	}

	retryWriteBatch := func(batch []*DeleteData) {
		if len(batch) == 0 {
			return
		}
		for {
			err := ds.writeBatch(batch)
			if err != nil {
				ds.errorChan <- fmt.Errorf("couldnt write batch: %w", err)
				time.Sleep(5 * time.Second)
			} else {
				return
			}
		}
	}

	for {
		select {
		case data, ok := <-ds.fanInChan:
			if !ok {
				retryWriteBatch(batch)
				close(ds.mayShutdown)
				return
			}
			batch = append(batch, data)
			if len(batch) == ds.batchSize {
				retryWriteBatch(batch)
				batch = clearBatch(batch)
			}
		case <-ticker.C:
			retryWriteBatch(batch)
			batch = clearBatch(batch)
		}
	}
}

func (ds *DefaultDeleter) writeBatch(batch []*DeleteData) error {
	uids := make([]string, 0)
	shortIDs := make([]string, 0)
	for _, v := range batch {
		if v != nil {
			uids = append(uids, v.uid)
			shortIDs = append(shortIDs, v.shortID)
		}
	}
	i := 1
	uidsPlacement := ""
	first := true
	for ; i <= len(uids); i++ {
		if first {
			uidsPlacement += fmt.Sprintf("$%d", i)
			first = false
		} else {
			uidsPlacement += fmt.Sprintf(",$%d", i)
		}

	}
	first = true
	shortIDsPlacement := ""
	for ; i <= 2*len(shortIDs); i++ {
		if first {
			shortIDsPlacement += fmt.Sprintf("$%d", i)
			first = false
		} else {
			shortIDsPlacement += fmt.Sprintf(",$%d", i)
		}
	}
	values := make([]interface{}, 0, 2*len(uids))
	for _, v := range uids {
		values = append(values, v)
	}
	for _, v := range shortIDs {
		values = append(values, v)
	}

	_, err := ds.db.Exec(fmt.Sprintf(`UPDATE records SET is_deleted=true
				FROM (
					SELECT records.id FROM records
					JOIN users ON records.id=users.record_id
					WHERE users.uid IN (%s) AND records.short_id IN (%s)
				) as d
				WHERE records.id=d.id;`, uidsPlacement, shortIDsPlacement), values...)
	if err != nil {
		e := fmt.Errorf("error update is_deleted: %w", err)
		ds.errorChan <- e
		return e
	}
	return nil
}
