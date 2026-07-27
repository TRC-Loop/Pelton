package desktop

import (
	"testing"

	"github.com/TRC-Loop/Pelton/internal/storage"
)

func TestMessageRemoteKey(t *testing.T) {
	// a present Message-ID is used verbatim (lowercased/trimmed) so the allow
	// survives re-sync, expunge and a changed local row id.
	got := messageRemoteKey(&storage.Message{ID: 7, MessageID: "  <ABC@Example.com>  "})
	if got != "<abc@example.com>" {
		t.Errorf("messageRemoteKey with Message-ID = %q, want %q", got, "<abc@example.com>")
	}
	// without a Message-ID it falls back to the local row id.
	if got := messageRemoteKey(&storage.Message{ID: 42}); got != "local:42" {
		t.Errorf("messageRemoteKey fallback = %q, want %q", got, "local:42")
	}
}
