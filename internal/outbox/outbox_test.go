package outbox

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

// newTestQueue opens a real migrated store in a temp dir and seeds one account,
// since the outbox foreign-keys accounts. It exercises the 0006 migration too.
func newTestQueue(t *testing.T) (*Queue, int64) {
	t.Helper()
	ctx := context.Background()

	db, err := storage.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.RunMigrations(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	id, err := db.CreateAccount(ctx, &storage.Account{Email: "me@example.com"})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	return NewQueue(db), id
}

func sampleMessage(accountID int64) Message {
	return Message{
		AccountID:    accountID,
		EnvelopeFrom: "me@example.com",
		Recipients:   []string{"a@example.com", "b@example.com"},
		Raw:          []byte("Subject: hi\r\n\r\nbody\r\n"),
	}
}

// fakeTransmitter records calls and fails a configurable number of times first.
type fakeTransmitter struct {
	failFor int
	calls   int
}

func (f *fakeTransmitter) Transmit(ctx context.Context, m Message) error {
	f.calls++
	if f.calls <= f.failFor {
		return errors.New("simulated transient failure")
	}
	return nil
}

func TestEnqueueAndSendSuccess(t *testing.T) {
	ctx := context.Background()
	q, accountID := newTestQueue(t)

	id, err := q.Enqueue(ctx, sampleMessage(accountID))
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	tx := &fakeTransmitter{}
	worker := NewWorker(q, tx)
	if err := worker.DrainOnce(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if tx.calls != 1 {
		t.Fatalf("transmit calls = %d, want 1", tx.calls)
	}

	got := findByID(t, q, id)
	if got.State != StateSent {
		t.Fatalf("state = %q, want %q", got.State, StateSent)
	}
	// recipients survive the storage round trip, including the second one.
	if len(got.Recipients) != 2 {
		t.Fatalf("recipients = %v, want 2", got.Recipients)
	}
}

func TestFailedAttemptRequeuesWithBackoff(t *testing.T) {
	ctx := context.Background()
	q, accountID := newTestQueue(t)

	id, err := q.Enqueue(ctx, sampleMessage(accountID))
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	// fail every time so we can inspect the requeue.
	worker := NewWorker(q, &fakeTransmitter{failFor: 100})
	if err := worker.DrainOnce(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}

	got := findByID(t, q, id)
	if got.State != StateQueued {
		t.Fatalf("state = %q, want %q (requeued)", got.State, StateQueued)
	}
	if got.Attempts != 1 {
		t.Fatalf("attempts = %d, want 1", got.Attempts)
	}
	if got.LastError == "" {
		t.Fatal("last error should be retained")
	}
	// backoff pushes the next attempt into the future, so a second drain now is a
	// no-op (nothing due).
	if !got.NextAttemptAt.After(time.Now().UTC()) {
		t.Fatalf("next attempt %v should be in the future", got.NextAttemptAt)
	}
}

func TestMessageFailsAfterMaxAttempts(t *testing.T) {
	ctx := context.Background()
	q, accountID := newTestQueue(t)

	id, err := q.Enqueue(ctx, sampleMessage(accountID))
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	// drive attempts directly so we do not have to wait out the backoff windows.
	for i := range MaxAttempts {
		m := findByID(t, q, id)
		retry, err := q.markAttemptFailed(ctx, m, "boom")
		if err != nil {
			t.Fatalf("mark attempt failed: %v", err)
		}
		if i < MaxAttempts-1 && !retry {
			t.Fatalf("attempt %d should still retry", i+1)
		}
		if i == MaxAttempts-1 && retry {
			t.Fatal("final attempt should not retry")
		}
	}

	got := findByID(t, q, id)
	if got.State != StateFailed {
		t.Fatalf("state = %q, want %q", got.State, StateFailed)
	}
	if got.LastError != "boom" {
		t.Fatalf("last error = %q, want retained cause", got.LastError)
	}
}

func TestRequeueStuckRecoversSendingRows(t *testing.T) {
	ctx := context.Background()
	q, accountID := newTestQueue(t)

	id, err := q.Enqueue(ctx, sampleMessage(accountID))
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	// claim it (moves to sending) but never finish, simulating a crash.
	if _, err := q.claimDue(ctx, time.Now().UTC()); err != nil {
		t.Fatalf("claim: %v", err)
	}
	if got := findByID(t, q, id); got.State != StateSending {
		t.Fatalf("state = %q, want %q", got.State, StateSending)
	}

	n, err := q.RequeueStuck(ctx)
	if err != nil {
		t.Fatalf("requeue stuck: %v", err)
	}
	if n != 1 {
		t.Fatalf("recovered = %d, want 1", n)
	}
	if got := findByID(t, q, id); got.State != StateQueued {
		t.Fatalf("state = %q, want %q", got.State, StateQueued)
	}
}

func TestBackoffIsCappedAndExponential(t *testing.T) {
	if backoff(1) != baseBackoff {
		t.Fatalf("backoff(1) = %v, want %v", backoff(1), baseBackoff)
	}
	if backoff(2) != 2*baseBackoff {
		t.Fatalf("backoff(2) = %v, want %v", backoff(2), 2*baseBackoff)
	}
	if backoff(100) != maxBackoff {
		t.Fatalf("backoff(100) = %v, want cap %v", backoff(100), maxBackoff)
	}
}

func findByID(t *testing.T, q *Queue, id int64) Message {
	t.Helper()
	rows, err := q.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, r := range rows {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("message %d not found", id)
	return Message{}
}
