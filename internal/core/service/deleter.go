package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
)

type DeleterService struct {
	wg          sync.WaitGroup
	repo        ports.DeleterRepository
	logger      ports.Logger
	batchSize   int
	mayShutdown chan struct{}

	errorChan chan error
	fanInChan chan *domain.DeleteData
}

func NewDeleter(repo ports.DeleterRepository, logger ports.Logger, batchSize int) (*DeleterService, error) {
	service := &DeleterService{
		repo:        repo,
		logger:      logger,
		batchSize:   batchSize,
		errorChan:   make(chan error, batchSize),
		fanInChan:   make(chan *domain.DeleteData, batchSize),
		mayShutdown: make(chan struct{}),
	}
	return service, nil
}

var _ ports.DeleterService = (*DeleterService)(nil)

func (ds *DeleterService) Run() error {
	// запустить горутину, которая пишет в базу
	// запустить горутину, которая логирует ошибки
	//
	go ds.dbWriter()
	go ds.errorLogger()

	return nil
}

func (ds *DeleterService) Process(_ context.Context, uid string, shortIDs []string) {
	ds.wg.Add(1)
	go ds.generator(uid, shortIDs)
}

func (ds *DeleterService) Shutdown() error {
	ds.wg.Wait()
	close(ds.fanInChan)
	<-ds.mayShutdown
	return nil
}

func (ds *DeleterService) generator(uid string, shortIDs []string) {
	for _, v := range shortIDs {
		ds.fanInChan <- &domain.DeleteData{
			UID:     uid,
			ShortID: v,
		}
	}
	ds.wg.Done()
}

func (ds *DeleterService) errorLogger() {
	for err := range ds.errorChan {
		ds.logger.Errorln("write batch err: ", err)
	}
	close(ds.mayShutdown)
}

func (ds *DeleterService) dbWriter() {
	batch := make([]*domain.DeleteData, 0, ds.batchSize)
	ticker := time.NewTicker(time.Second)

	clearBatch := func(batch []*domain.DeleteData) []*domain.DeleteData {
		for i := range batch {
			batch[i] = nil
		}
		return batch[:0]
	}

	retryWriteBatch := func(batch []*domain.DeleteData) {
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
				close(ds.errorChan)
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

func (ds *DeleterService) writeBatch(batch []*domain.DeleteData) error {
	err := ds.repo.Delete(context.Background(), batch)
	if err != nil {
		ds.errorChan <- err
		return err
	}
	return nil
}