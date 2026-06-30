// Command imaptest exercises internal/imap end-to-end against a real server:
// connect, login, list folders, read INBOX headers, fetch one message, toggle
// a flag, then idle until Ctrl+C. Credentials come from the environment:
//
//	IMAP_HOST, IMAP_PORT, IMAP_USER, IMAP_PASSWORD
//	IMAP_INSECURE=1   skip TLS verification (debug only)
//	IMAP_DEBUG=1      dump the raw protocol to stderr (leaks credentials)
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/emersion/go-imap/v2"

	pimap "github.com/TRC-Loop/Pelton/internal/imap"
)

const (
	inboxName    = "INBOX"
	headerLimit  = 10
	snippetLen   = 500
	flagToToggle = imap.FlagFlagged // safe and reversible
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := configFromEnv()
	if err != nil {
		return err
	}

	fmt.Printf("connecting to %s:%d ...\n", cfg.Host, cfg.Port)
	client, err := pimap.Connect(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Login(); err != nil {
		return err
	}
	fmt.Printf("logged in as %s\n", cfg.Username)
	// best-effort logout on the way out
	defer func() {
		if err := client.Logout(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", err)
		}
	}()

	if err := listFolders(client); err != nil {
		return err
	}

	mbox, err := client.Select(inboxName)
	if err != nil {
		return err
	}
	fmt.Printf("\n=== %s ===\n", mbox.Name)
	fmt.Printf("messages: %d  uidnext: %d  uidvalidity: %d\n",
		mbox.NumMessages, mbox.UIDNext, mbox.UIDValidity)

	headers, err := printHeaders(client)
	if err != nil {
		return err
	}

	if len(headers) == 0 {
		fmt.Println("\ninbox is empty; skipping message, flag and idle demos")
		return nil
	}

	newest := headers[0]
	if err := printMessage(client, newest.UID); err != nil {
		return err
	}
	if err := flagDemo(client, newest.UID); err != nil {
		return err
	}
	return idleDemo(client)
}

func configFromEnv() (pimap.Config, error) {
	cfg := pimap.Config{
		Host:               os.Getenv("IMAP_HOST"),
		Username:           os.Getenv("IMAP_USER"),
		Password:           os.Getenv("IMAP_PASSWORD"),
		InsecureSkipVerify: os.Getenv("IMAP_INSECURE") == "1",
	}
	if cfg.Host == "" || cfg.Username == "" || cfg.Password == "" {
		return cfg, fmt.Errorf("IMAP_HOST, IMAP_USER and IMAP_PASSWORD must be set")
	}

	if p := os.Getenv("IMAP_PORT"); p != "" {
		port, err := strconv.Atoi(p)
		if err != nil {
			return cfg, fmt.Errorf("invalid IMAP_PORT %q: %w", p, err)
		}
		cfg.Port = port
	}
	if os.Getenv("IMAP_DEBUG") == "1" {
		cfg.DebugWriter = os.Stderr
	}
	return cfg, nil
}

func listFolders(client *pimap.Client) error {
	folders, err := client.ListFolders()
	if err != nil {
		return err
	}
	fmt.Printf("\n=== folders (%d) ===\n", len(folders))
	for _, f := range folders {
		marker := " "
		if !f.Selectable() {
			marker = "x" // not selectable (container)
		}
		fmt.Printf(" [%s] %-30s delim=%q %s\n", marker, f.Name, f.Delimiter, attrString(f.Attrs))
	}
	return nil
}

func printHeaders(client *pimap.Client) ([]pimap.MessageHeader, error) {
	headers, err := client.FetchRecentHeaders(headerLimit)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n=== %d most recent headers ===\n", len(headers))
	for _, h := range headers {
		fmt.Printf("uid %-6d %s\n", h.UID, h.Date.Format(time.RFC1123Z))
		fmt.Printf("  from:    %s\n", h.From)
		fmt.Printf("  subject: %s\n", h.Subject)
		fmt.Printf("  flags:   %s\n", flagString(h.Flags))
	}
	return headers, nil
}

func printMessage(client *pimap.Client, uid imap.UID) error {
	msg, err := client.FetchMessage(uid)
	if err != nil {
		return err
	}
	fmt.Printf("\n=== full message uid %d ===\n", msg.UID)
	fmt.Printf("from:    %s\n", msg.From)
	fmt.Printf("to:      %s\n", msg.To)
	fmt.Printf("date:    %s\n", msg.Date.Format(time.RFC1123Z))
	fmt.Printf("subject: %s\n", msg.Subject)
	fmt.Printf("html:    %v\n", msg.HTML != "")
	if len(msg.Attachments) > 0 {
		fmt.Printf("attachments (%d):\n", len(msg.Attachments))
		for _, a := range msg.Attachments {
			fmt.Printf("  - %s (%s)\n", a.Filename, a.ContentType)
		}
	}
	fmt.Printf("text:\n%s\n", snippet(msg.Text, snippetLen))
	return nil
}

// flagDemo adds flagToToggle, reads it back, then removes it, printing flags
// before and after each step. Fully reversible.
func flagDemo(client *pimap.Client, uid imap.UID) error {
	fmt.Printf("\n=== flag demo on uid %d (%s) ===\n", uid, flagToToggle)

	before, err := client.FetchFlags(uid)
	if err != nil {
		return err
	}
	fmt.Printf("before: %s\n", flagString(before))

	if err := client.AddFlags(uid, flagToToggle); err != nil {
		return err
	}
	added, err := client.FetchFlags(uid)
	if err != nil {
		return err
	}
	fmt.Printf("added:  %s\n", flagString(added))

	if err := client.RemoveFlags(uid, flagToToggle); err != nil {
		return err
	}
	restored, err := client.FetchFlags(uid)
	if err != nil {
		return err
	}
	fmt.Printf("after:  %s\n", flagString(restored))
	return nil
}

// idleDemo idles until SIGINT/SIGTERM, printing pushes as they arrive.
func idleDemo(client *pimap.Client) error {
	if !client.SupportsIdle() {
		fmt.Println("\nserver does not support IDLE; skipping idle demo")
		return nil
	}

	fmt.Println("\n=== idle: waiting for new mail, press Ctrl+C to stop ===")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case u := <-client.Updates():
				switch {
				case u.NumMessages != nil:
					fmt.Printf("** new mail: mailbox now has %d messages\n", *u.NumMessages)
				case u.ExpungedSeqNum != nil:
					fmt.Printf("** message expunged: seq %d\n", *u.ExpungedSeqNum)
				}
			}
		}
	}()

	if err := client.Idle(ctx); err != nil {
		return err
	}
	fmt.Println("idle stopped cleanly")
	return nil
}
