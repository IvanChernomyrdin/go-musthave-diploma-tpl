package listener

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type WorkerPool struct {
	workers       int
	jobChan       chan Job
	workerFunc    func(context.Context, Job) error
	logger        *zap.SugaredLogger
	wg            sync.WaitGroup
	batchSize     int
	jobBuffer     []Job
	bufferMutex   sync.RWMutex
	bufferFull    chan struct{}
	maxBufferSize int
}

func NewWorkerPool(workers, batchSize, maxBufferSize int, workerFunc func(context.Context, Job) error, logger *zap.SugaredLogger) *WorkerPool {
	return &WorkerPool{
		workers:       workers,
		jobChan:       make(chan Job, workers*10),
		workerFunc:    workerFunc,
		logger:        logger,
		batchSize:     batchSize,
		jobBuffer:     make([]Job, 0, batchSize),
		bufferFull:    make(chan struct{}, 1),
		maxBufferSize: maxBufferSize,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	// Запускаем воркеры
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}

	// Запускаем обработчик буфера
	wp.wg.Add(1)
	go wp.bufferProcessor(ctx)
}

func (wp *WorkerPool) bufferProcessor(ctx context.Context) {
	defer wp.wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			wp.flushBuffer()
			return
		case <-wp.bufferFull:
			wp.flushBuffer()
		case <-ticker.C:
			wp.flushBufferIfNeeded()
		}
	}
}

func (wp *WorkerPool) flushBufferIfNeeded() {
	wp.bufferMutex.Lock()
	defer wp.bufferMutex.Unlock()

	if len(wp.jobBuffer) >= wp.batchSize {
		wp.flushBufferLocked()
	}
}

func (wp *WorkerPool) flushBuffer() {
	wp.bufferMutex.Lock()
	defer wp.bufferMutex.Unlock()

	if len(wp.jobBuffer) > 0 {
		wp.flushBufferLocked()
	}
}

func (wp *WorkerPool) flushBufferLocked() {
	for _, job := range wp.jobBuffer {
		select {
		case wp.jobChan <- job:
		default:
			wp.logger.Warn("Worker pool job channel full, dropping job from buffer")
		}
	}

	wp.jobBuffer = make([]Job, 0, wp.batchSize)
	wp.logger.Debugf("Flushed buffer, sent %d jobs to workers", len(wp.jobBuffer))
}

func (wp *WorkerPool) Submit(job Job) {
	wp.bufferMutex.Lock()
	defer wp.bufferMutex.Unlock()

	if len(wp.jobBuffer) >= wp.maxBufferSize {
		wp.logger.Warn("Buffer size exceeded, forcing flush")
		wp.flushBufferLocked()
	}

	wp.jobBuffer = append(wp.jobBuffer, job)
	wp.logger.Debugf("Job added to buffer, current buffer size: %d", len(wp.jobBuffer))

	if len(wp.jobBuffer) >= wp.batchSize {
		select {
		case wp.bufferFull <- struct{}{}:
		default:
		}
	}
}

func (wp *WorkerPool) SubmitBatch(jobs []Job) {
	if len(jobs) == 0 {
		return
	}

	wp.bufferMutex.Lock()
	defer wp.bufferMutex.Unlock()

	if len(wp.jobBuffer)+len(jobs) > wp.maxBufferSize {
		wp.logger.Warn("Batch too large, flushing buffer first")
		wp.flushBufferLocked()
	}

	wp.jobBuffer = append(wp.jobBuffer, jobs...)
	wp.logger.Debugf("Batch of %d jobs added to buffer, total size: %d", len(jobs), len(wp.jobBuffer))

	if len(wp.jobBuffer) >= wp.batchSize {
		select {
		case wp.bufferFull <- struct{}{}:
		default:
		}
	}
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()

	wp.logger.Infof("Worker %d started", id)

	for {
		select {
		case job, ok := <-wp.jobChan:
			if !ok {
				wp.logger.Infof("Worker %d stopping: channel closed", id)
				return
			}

			wp.logger.Debugf("Worker %d processing job: OrderID=%d, Number=%s", id, job.OrderID, job.Number)

			if err := wp.workerFunc(ctx, job); err != nil {
				wp.logger.Errorf("Worker %d failed to process job: %v", id, err)

				// Повторная попытка с экспоненциальной задержкой
				if job.Attempt < 3 {
					job.Attempt++
					retryDelay := time.Duration(job.Attempt*job.Attempt) * time.Second

					wp.logger.Infof("Worker %d retrying job in %v (attempt %d)", id, retryDelay, job.Attempt)

					time.AfterFunc(retryDelay, func() {
						wp.Submit(job)
					})
				} else {
					wp.logger.Errorf("Worker %d max retries exceeded for job: %+v", id, job)
				}
			} else {
				wp.logger.Debugf("Worker %d successfully processed job OrderID=%d", id, job.OrderID)
			}

		case <-ctx.Done():
			wp.logger.Infof("Worker %d stopping: context cancelled", id)
			return
		}
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.jobChan)
	wp.wg.Wait()
	wp.logger.Info("Worker pool stopped")
}

func (wp *WorkerPool) GetBufferSize() int {
	wp.bufferMutex.RLock()
	defer wp.bufferMutex.RUnlock()
	return len(wp.jobBuffer)
}
