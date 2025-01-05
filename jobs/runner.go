package jobs

import (
	"context"
	"sync"
	"time"

	"cyberix.fr/frcc/messaging"
	"cyberix.fr/frcc/models"
	"go.uber.org/zap"
)

type Func = func(context.Context, models.Message) error

type Runner struct {
	emailer *messaging.Emailer
	jobs    map[string]Func
	log     *zap.Logger
	queue   *messaging.Queue
}

type NewRunnerOptions struct {
	Emailer *messaging.Emailer
	Log     *zap.Logger
	Queue   *messaging.Queue
}

func NewRunner(opts NewRunnerOptions) *Runner {
	if opts.Log == nil {
		opts.Log = zap.NewNop()
	}

	return &Runner{
		emailer: opts.Emailer,
		jobs:    map[string]Func{},
		log:     opts.Log,
		queue:   opts.Queue,
	}
}

func (r *Runner) Start(ctx context.Context) {
	r.log.Info("Starting")
	r.registerJobs()
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			r.log.Info("Stopping")
			wg.Wait()
			return
		default:
			r.receiveAndRun(ctx, &wg)
		}
	}
}

func (r *Runner) receiveAndRun(ctx context.Context, wg *sync.WaitGroup) {
	m, receiptID, err := r.queue.Receive(ctx)
	if err != nil {
		r.log.Info("Error receiving message", zap.Error(err))
		time.Sleep(time.Second)
		return
	}

	if m == nil {
		return
	}

	name, ok := (*m)["job"]
	if !ok {
		r.log.Info("Error getting job name from message")
		return
	}

	job, ok := r.jobs[name]
	if !ok {
		r.log.Info("No job with this name", zap.String("name", name))
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		log := r.log.With(zap.String("name", name))

		defer func() {
			if rec := recover(); rec != nil {
				log.Info("Recovered from panic in job", zap.Any("recover", rec))
			}
		}()

		before := time.Now()
		if err := job(ctx, *m); err != nil {
			log.Info("Error running job", zap.Error(err))
			return
		}
		after := time.Now()
		duration := after.Sub(before)
		log.Info("Successfully ran job", zap.Duration("duration", duration))

		deleteCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := r.queue.Delete(deleteCtx, receiptID); err != nil {
			log.Info("Error deleting message, job will be repeated", zap.Error(err))
		}
	}()
}

type registry interface {
	Register(name string, fn Func)
}

func (r *Runner) Register(name string, j Func) {
	r.jobs[name] = j
}
