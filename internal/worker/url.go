package worker

import (
	"context"
	"time"

	"github.com/JustTony97/url-shortener.git/internal/logger"
	"github.com/JustTony97/url-shortener.git/internal/model"
)

type URLRepository interface {
	DeleteUserURLs(ctx context.Context, userIDs []string, shortURLs []string) error
}

type URLBatchWorker struct {
	tasksChan <-chan model.DeleteURLTask
	batchSize int
	timeout   time.Duration
	repo      URLRepository
}

func NewURLBatchWorker(tasksChan <-chan model.DeleteURLTask, batchSize int, timeout time.Duration, repo URLRepository) *URLBatchWorker {
	return &URLBatchWorker{
		tasksChan: tasksChan,
		batchSize: batchSize,
		timeout:   timeout,
		repo:      repo,
	}
}

func (bw *URLBatchWorker) flushBatch(batch []model.DeleteURLTask) {
	logger.Log.Info("Flushing butch!")
	totalLines := 0
	for _, t := range batch {
		totalLines += len(t.ShortURLs)
	}

	userIDs := make([]string, 0, totalLines)
	shortURLs := make([]string, 0, totalLines)

	for _, task := range batch {
		for _, url := range task.ShortURLs {
			userIDs = append(userIDs, task.UserID)
			shortURLs = append(shortURLs, url)
		}
	}
	_ = bw.repo.DeleteUserURLs(context.Background(), userIDs, shortURLs)
}

func (bw *URLBatchWorker) StartWorker() {
	batch := make([]model.DeleteURLTask, 0, bw.batchSize)

	ticker := time.NewTicker(bw.timeout)
	defer ticker.Stop()

	for {
		select {
		case task, ok := <-bw.tasksChan:
			if !ok {
				bw.flushBatch(batch)
				return
			}
			batch = append(batch, task)

			if len(batch) >= bw.batchSize {
				bw.flushBatch(batch)
				batch = make([]model.DeleteURLTask, 0, bw.batchSize)
				ticker.Reset(bw.timeout)

			}
		case <-ticker.C:
			if len(batch) > 0 {
				bw.flushBatch(batch)
				batch = make([]model.DeleteURLTask, 0, bw.batchSize)
			}
		}
	}
}
