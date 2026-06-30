package outbox

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// defaultPollInterval is how often the worker checks for newly due messages when
// the queue is idle. A new enqueue is picked up within this window.
const defaultPollInterval = 5 * time.Second

// Transmitter sends one fully built message. It is implemented by the smtp
// sender, which transmits over SMTP and appends to the Sent folder. Defining it
// here keeps the worker decoupled from the smtp and imap packages.
type Transmitter interface {
	Transmit(ctx context.Context, m Message) error
}

// Worker drains the queue in the background. It is cancellable via context and
// survives transient transmit errors by leaving the message queued for a later
// attempt, so queued mail is never lost to a flaky network.
type Worker struct {
	queue       *Queue
	transmitter Transmitter
	log         *slog.Logger
	interval    time.Duration
	onChange    func()
}

// Option configures a Worker.
type Option func(*Worker)

// WithLogger sets the worker's logger.
func WithLogger(log *slog.Logger) Option {
	return func(w *Worker) { w.log = log }
}

// WithOnChange registers a callback fired whenever a message's state changes
// (claimed, sent or failed). The app uses it to push an outbox-changed event to
// the ui after the new state is persisted, so the ui never gets stuck showing a
// stale "sending". The callback must not block.
func WithOnChange(fn func()) Option {
	return func(w *Worker) {
		if fn != nil {
			w.onChange = fn
		}
	}
}

// WithInterval overrides the idle poll interval.
func WithInterval(d time.Duration) Option {
	return func(w *Worker) {
		if d > 0 {
			w.interval = d
		}
	}
}

// NewWorker builds a worker for the queue and transmitter.
func NewWorker(queue *Queue, transmitter Transmitter, opts ...Option) *Worker {
	w := &Worker{
		queue:       queue,
		transmitter: transmitter,
		log:         slog.New(slog.DiscardHandler),
		interval:    defaultPollInterval,
		onChange:    func() {},
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// Run drains the queue until ctx is cancelled, polling for new work in between.
// It first recovers any messages stranded in the sending state by a prior crash.
// It returns ctx.Err() on cancellation.
func (w *Worker) Run(ctx context.Context) error {
	if n, err := w.queue.RequeueStuck(ctx); err != nil {
		w.log.Warn("could not requeue stranded messages", "err", err)
	} else if n > 0 {
		w.log.Info("requeued stranded messages", "count", n)
		w.onChange()
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		if err := w.DrainOnce(ctx); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			// a claim error is unusual (db level); log and keep the worker alive.
			w.log.Error("outbox drain error", "err", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// DrainOnce sends every message that is currently due, then returns. Useful for
// tests and for a one-shot flush. A transmit failure on one message does not
// stop the others; it is recorded and the message is requeued or failed.
func (w *Worker) DrainOnce(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		m, err := w.queue.claimDue(ctx, time.Now().UTC())
		if err != nil {
			return err
		}
		if m == nil {
			return nil
		}
		// the row is now in the sending state; let the ui reflect that, then
		// reflect the final outcome once process records it.
		w.onChange()
		w.process(ctx, *m)
		w.onChange()
	}
}

// process transmits one claimed message and records the outcome.
func (w *Worker) process(ctx context.Context, m Message) {
	err := w.transmitter.Transmit(ctx, m)
	if err == nil {
		if e := w.queue.markSent(ctx, m.ID); e != nil {
			w.log.Error("failed to mark message sent", "id", m.ID, "err", e)
			return
		}
		w.log.Info("message sent", "id", m.ID, "to", m.Recipients)
		return
	}

	retry, e := w.queue.markAttemptFailed(ctx, m, err.Error())
	if e != nil {
		w.log.Error("failed to record send failure", "id", m.ID, "err", e)
		return
	}
	if retry {
		w.log.Warn("send failed, will retry", "id", m.ID, "attempt", m.Attempts+1, "err", err)
	} else {
		w.log.Error("send failed permanently", "id", m.ID, "attempts", m.Attempts+1, "err", err)
	}
}
