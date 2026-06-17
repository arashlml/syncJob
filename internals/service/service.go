package service

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

type Iterator interface {
	HasNext(cursor int) bool
	Next(cursor int) (map[string]interface{}, error)
}

type Insert interface {
	Insert(data map[string]interface{}) error
}

type SyncService struct {
	syncChannel     chan map[string]interface{}
	chanSize        int
	numberOfWorkers int
	iterator        Iterator
	insertFunction  Insert
	cursor          int
	wg              *sync.WaitGroup
	logger          *slog.Logger
}

func NewSyncService(iterator Iterator, insert Insert, numberOfWorkers int, chanSize int, serviceName string) *SyncService {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		slog.String("service", serviceName),
		slog.String("component", "SyncService"),
	)

	return &SyncService{
		syncChannel:     make(chan map[string]interface{}, chanSize),
		chanSize:        chanSize,
		iterator:        iterator,
		insertFunction:  insert,
		numberOfWorkers: numberOfWorkers,
		wg:              &sync.WaitGroup{},
		logger:          logger,
	}
}

func (s *SyncService) Start() {
	s.logger.Info("starting sync service",
		slog.Int("workers", s.numberOfWorkers),
		slog.Int("channel_size", s.chanSize),
	)
	s.wg.Add(s.numberOfWorkers)
	for i := 0; i < s.numberOfWorkers; i++ {
		go s.Consume()
	}
	s.wg.Add(1)
	go s.Read()
}

func (s *SyncService) Wait() {
	s.wg.Wait()
	s.logger.Info("all workers finished")
}

func (s *SyncService) Read() {
	defer s.wg.Done()
	defer close(s.syncChannel)

	s.logger.Info("reader started", slog.String("reason", "iterating over source data"))

	for {
		if !s.iterator.HasNext(s.cursor) {
			s.logger.Info("reader finished", slog.Int("total_read", s.cursor))
			return
		}
		data, err := s.iterator.Next(s.cursor)
		if err != nil {
			s.logger.Error("failed to read next item",
				slog.String("reason", "iterator returned error"),
				slog.Int("cursor", s.cursor),
				slog.Any("error", err),
			)
			continue
		}
		s.cursor++
		s.syncChannel <- data
	}
}

func (s *SyncService) Consume() {
	defer s.wg.Done()

	s.logger.Info("worker started", slog.String("reason", "ready to consume from channel"))

	for data := range s.syncChannel {
		func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("panic recovered",
						slog.String("reason", "insert function panicked"),
						slog.Any("panic", r),
					)
				}
			}()
			if err := s.insertFunction.Insert(data); err != nil {
				s.logger.Error("failed to insert data",
					slog.String("reason", "insert function returned error"),
					slog.Any("error", err),
					slog.Any("id", data["id"]),
				)
				return
			}
			s.logger.Debug("item inserted successfully",
				slog.Any("id", data["id"]),
			)
		}()
	}

	s.logger.Info("worker finished", slog.String("reason", "channel closed"))
}

func (s *SyncService) WithContext(ctx context.Context, traceID string) {
	s.logger = s.logger.With(
		slog.String("trace_id", traceID),
	)
}
