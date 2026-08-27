package carddav

import (
	"strings"
	"testing"
)

// a card written by another client, carrying properties Pelton knows nothing
// about. Editing a contact must not be what deletes them.
const foreignCard = `BEGIN:VCARD
VERSION:3.0
UID:abc-123
FN:Arne Kock
N:Kock;Arne;;;
EMAIL;TYPE=WORK:arne@example.com
EMAIL;TYPE=HOME:arne@home.example
TEL;TYPE=CELL:+49123456
ORG:Stellar Foundry
TITLE:Maker of potatoes
BDAY:1990-04-01
X-SOCIALPROFILE;TYPE=mastodon:https://example.social/@arne
END:VCARD
`

func TestFromCardReadsTheFieldsWeShow(t *testing.T) {
	contact, err := ParseCard(foreignCard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if contact.UID != "abc-123" {
		t.Errorf("UID = %q", contact.UID)
	}
	if contact.FullName != "Arne Kock" {
		t.Errorf("FullName = %q", contact.FullName)
	}
	if contact.Organization != "Stellar Foundry" || contact.Title != "Maker of potatoes" {
		t.Errorf("org = %q, title = %q", contact.Organization, contact.Title)
	}
	if len(contact.Emails) != 2 || contact.Emails[0].Value != "arne@example.com" {
		t.Fatalf("emails = %+v", contact.Emails)
	}
	if contact.Emails[0].Label != "work" {
		t.Errorf("first email label = %q, want work", contact.Emails[0].Label)
	}
	if contact.PrimaryEmail() != "arne@example.com" {
		t.Errorf("PrimaryEmail = %q", contact.PrimaryEmail())
	}
	if len(contact.Phones) != 1 || contact.Phones[0].Value != "+49123456" {
		t.Errorf("phones = %+v", contact.Phones)
	}
}

// The round trip is the one that matters: an edit here has to leave everything
// else on the card alone, because it belongs to whoever wrote it.
func TestToCardKeepsPropertiesItDoesNotUnderstand(t *testing.T) {
	contact, err := ParseCard(foreignCard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	contact.Title = "Potato farmer"
	contact.Phones = nil

	encoded, err := EncodeCard(contact.ToCard())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	for _, want := range []string{"BDAY:1990-04-01", "X-SOCIALPROFILE", "UID:abc-123", "Potato farmer"} {
		if !strings.Contains(encoded, want) {
			t.Errorf("encoded card lost %q:\n%s", want, encoded)
		}
	}
	if strings.Contains(encoded, "+49123456") {
		t.Errorf("cleared phone survived:\n%s", encoded)
	}
	if strings.Contains(encoded, "Maker of potatoes") {
		t.Errorf("old title survived:\n%s", encoded)
	}
}

// clearing a single-valued field has to remove the property, not write it
// empty, or the server keeps showing the old value in some clients.
func TestClearingAFieldRemovesIt(t *testing.T) {
	contact, err := ParseCard(foreignCard)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	contact.Organization = ""

	encoded, err := EncodeCard(contact.ToCard())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if strings.Contains(encoded, "ORG") {
		t.Errorf("cleared org still present:\n%s", encoded)
	}
}

// a card with no FN is legal and unusable in a list, so something readable has
// to stand in for it.
func TestFullNameFallsBackToNameThenAddress(t *testing.T) {
	fromName, err := ParseCard("BEGIN:VCARD\nVERSION:3.0\nN:Kock;Arne;;;\nEMAIL:arne@example.com\nEND:VCARD\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if fromName.FullName != "Arne Kock" {
		t.Errorf("FullName = %q, want the structured name", fromName.FullName)
	}

	fromEmail, err := ParseCard("BEGIN:VCARD\nVERSION:3.0\nEMAIL:arne@example.com\nEND:VCARD\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if fromEmail.FullName != "arne@example.com" {
		t.Errorf("FullName = %q, want the address", fromEmail.FullName)
	}
}

// a contact created here has no card to start from.
func TestNewContactEncodes(t *testing.T) {
	contact := Contact{
		UID:      "new-1",
		FullName: "Spud McPelton",
		Emails:   []Labelled{{Value: "spud@example.com", Label: "home"}},
	}
	encoded, err := EncodeCard(contact.ToCard())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	for _, want := range []string{"FN:Spud McPelton", "spud@example.com", "UID:new-1", "N:"} {
		if !strings.Contains(encoded, want) {
			t.Errorf("new card missing %q:\n%s", want, encoded)
		}
	}
}
