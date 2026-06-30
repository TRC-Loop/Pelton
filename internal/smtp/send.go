package smtp

import (
	"context"
	"log/slog"
	"time"

	"github.com/TRC-Loop/Pelton/internal/crypto"
	"github.com/TRC-Loop/Pelton/internal/outbox"
)

// Enqueue builds the message (applying crypto when requested) and adds it to the
// outbox for the worker to transmit. Because BuildRaw hard-fails when crypto is
// requested but cannot be completed, an encrypted message is only ever enqueued
// in its encrypted form, and a protected message is never queued as plaintext.
//
// notBefore holds the message queued until that time (the undo-send delay). Pass
// the zero time to send as soon as the worker picks it up.
func Enqueue(ctx context.Context, queue *outbox.Queue, accountID int64, msg *Message, engine crypto.Engine, mode crypto.Mode, opts crypto.Options, notBefore time.Time) (int64, error) {
	raw, err := BuildRaw(msg, engine, mode, opts)
	if err != nil {
		return 0, err
	}
	return queue.Enqueue(ctx, outbox.Message{
		AccountID:    accountID,
		EnvelopeFrom: msg.From.Email,
		Recipients:   msg.Recipients(),
		Raw:          raw,
		NotBefore:    notBefore,
	})
}

// SentAppender appends an already-sent message to the Sent folder. It is
// satisfied by a closure over the imap client, keeping smtp from depending on a
// live imap connection. Returns the folder used.
type SentAppender func(raw []byte) (string, error)

// Sender transmits queued messages over SMTP and, on success, appends them to
// the Sent folder. It implements outbox.Transmitter so the worker can drive it.
type Sender struct {
	cfg        Config
	appendSent SentAppender
	log        *slog.Logger
}

// SenderOption configures a Sender.
type SenderOption func(*Sender)

// WithSentAppender wires the append-to-Sent step run after a successful send.
// Without it, the step is skipped (with a debug log) and sending still succeeds.
func WithSentAppender(appender SentAppender) SenderOption {
	return func(s *Sender) { s.appendSent = appender }
}

// WithLogger sets the sender's logger.
func WithLogger(log *slog.Logger) SenderOption {
	return func(s *Sender) { s.log = log }
}

// NewSender builds a Sender for one account's submission config.
func NewSender(cfg Config, opts ...SenderOption) *Sender {
	s := &Sender{cfg: cfg, log: slog.New(slog.DiscardHandler)}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Transmit dials, authenticates, sends the message, then appends it to Sent.
// The append is non-fatal: the mail is already delivered, so an append failure
// is logged as a warning and Transmit still returns success.
func (s *Sender) Transmit(ctx context.Context, m outbox.Message) error {
	client, err := Dial(s.cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Authenticate(); err != nil {
		return err
	}
	if err := client.Send(ctx, m.EnvelopeFrom, m.Recipients, m.Raw); err != nil {
		return err
	}

	s.appendToSent(m.Raw)
	return nil
}

// appendToSent runs the optional append-to-Sent step, never failing the send.
func (s *Sender) appendToSent(raw []byte) {
	if s.appendSent == nil {
		s.log.Debug("no sent appender configured, skipping append to Sent")
		return
	}
	folder, err := s.appendSent(raw)
	if err != nil {
		s.log.Warn("append to Sent failed, message was still sent", "err", err)
		return
	}
	s.log.Info("appended to Sent", "folder", folder)
}

// compile-time check that Sender satisfies the worker's contract.
var _ outbox.Transmitter = (*Sender)(nil)
