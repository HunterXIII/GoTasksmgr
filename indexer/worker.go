package indexer

import (
	"context"
	"strings"
	"tasksmgr/repo"
	"time"
)

type Worker struct {
	queue chan int
	repo  *repo.TaskRepository
}

func NewWorker(queue chan int, repo *repo.TaskRepository) *Worker {
	return &Worker{
		queue: queue,
		repo:  repo,
	}
}

func (w *Worker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case taskId := <-w.queue:
			w.process(ctx, taskId)
		}
	}
}

func (w *Worker) process(ctx context.Context, taskId int) {
	task, err := w.repo.GetById(ctx, taskId)
	if err != nil {
		return
	}

	count := len(strings.Split(task.Title, " "))
	w.repo.UpdateWordCount(ctx, task.Id, count)
	time.Sleep(5 * time.Second)
}
