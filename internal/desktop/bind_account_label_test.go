package desktop

import (
	"testing"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

// TestLocalLabelStaysOffOutgoingMail is the bug from #326: the sidebar name and
// the From name used to be one field, so calling a mailbox "work junk" told
// every recipient so.
func TestLocalLabelStaysOffOutgoingMail(t *testing.T) {
	app, db, ctx := exportTestApp(t)
	id, err := db.CreateAccount(ctx, &storage.Account{
		Email:         "me@example.com",
		DisplayName:   "Arne Kock",
		LocalLabel:    "work junk",
		UseLocalLabel: true,
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	msg, err := app.buildMessage(ComposeRequest{
		AccountID: id,
		To:        []AddressDTO{{Email: "someone@example.com"}},
		Subject:   "hello",
		Text:      "hello",
	})
	if err != nil {
		t.Fatalf("build message: %v", err)
	}
	if msg.From.Name != "Arne Kock" {
		t.Errorf("From name = %q, want the display name, not the local label", msg.From.Name)
	}
}

// TestUpdateAccountStoresLabel covers the round trip the settings ui makes:
// both fields go out to the frontend and both come back.
func TestUpdateAccountStoresLabel(t *testing.T) {
	app, db, ctx := exportTestApp(t)
	id, err := db.CreateAccount(ctx, &storage.Account{Email: "me@example.com", DisplayName: "Arne Kock"})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	dto, err := app.UpdateAccount(UpdateAccountRequest{
		ID:            id,
		DisplayName:   "Arne Kock",
		LocalLabel:    "work junk",
		UseLocalLabel: true,
		IMAPHost:      "imap.example.com",
		IMAPPort:      993,
		SMTPHost:      "smtp.example.com",
		SMTPPort:      465,
	})
	if err != nil {
		t.Fatalf("update account: %v", err)
	}
	if dto.DisplayName != "Arne Kock" || dto.LocalLabel != "work junk" || !dto.UseLocalLabel {
		t.Errorf("dto = %+v, want both names carried", dto)
	}

	stored, err := db.GetAccount(ctx, id)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if stored.Label() != "work junk" || stored.DisplayName != "Arne Kock" {
		t.Errorf("stored label %q, display name %q", stored.Label(), stored.DisplayName)
	}
}
