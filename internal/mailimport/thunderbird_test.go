package mailimport

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePrefLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{"string", `user_pref("mail.server.server1.hostname", "imap.example.com");`, "mail.server.server1.hostname", "imap.example.com", true},
		{"number", `user_pref("mail.server.server1.port", 993);`, "mail.server.server1.port", "993", true},
		{"bool", `user_pref("mail.server.server1.login_at_startup", true);`, "mail.server.server1.login_at_startup", "true", true},
		{"escaped quote", `user_pref("mail.identity.id1.fullName", "A \"Nick\" B");`, "mail.identity.id1.fullName", `A "Nick" B`, true},
		{"escaped backslash", `user_pref("mail.server.server1.directory", "C:\\Users\\me\\Mail");`, "mail.server.server1.directory", `C:\Users\me\Mail`, true},
		{"empty string", `user_pref("mail.identity.id1.organization", "");`, "mail.identity.id1.organization", "", true},
		{"comment", `// Mozilla User Preferences`, "", "", false},
		{"blank", ``, "", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key, value, ok := parsePrefLine(tc.line)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if key != tc.wantKey || value != tc.wantValue {
				t.Fatalf("got (%q, %q), want (%q, %q)", key, value, tc.wantKey, tc.wantValue)
			}
		})
	}
}

func TestAccountsFromPrefs(t *testing.T) {
	p := prefs{
		"mail.accountmanager.accounts":     "account1,account2,account3,account9",
		"mail.account.account1.server":     "server1",
		"mail.account.account1.identities": "id1",
		"mail.server.server1.type":         "imap",
		"mail.server.server1.hostname":     "imap.example.com",
		"mail.server.server1.port":         "993",
		"mail.server.server1.userName":     "login-name",
		"mail.identity.id1.useremail":      "me@example.com",
		"mail.identity.id1.fullName":       "Me",
		"mail.identity.id1.smtpServer":     "smtp1",
		"mail.smtpserver.smtp1.hostname":   "smtp.example.com",
		"mail.smtpserver.smtp1.port":       "587",

		// account2 is Local Folders, which has no server to import.
		"mail.account.account2.server": "server2",
		"mail.server.server2.type":     "none",

		// account3 falls back to the default smtp server and has no separate
		// login name, which must not be stored as one.
		"mail.account.account3.server":     "server3",
		"mail.account.account3.identities": "id3",
		"mail.server.server3.type":         "imap",
		"mail.server.server3.hostname":     "imap.other.example",
		"mail.server.server3.port":         "143",
		"mail.server.server3.userName":     "other@example.com",
		"mail.identity.id3.useremail":      "other@example.com",
		"mail.smtp.defaultserver":          "smtp1",

		// account9 is listed but its server entry is gone, the leftover of a
		// removal. It must not produce an account with no host.
	}

	accounts := accountsFromPrefs(p)
	if len(accounts) != 2 {
		t.Fatalf("got %d accounts, want 2: %+v", len(accounts), accounts)
	}

	first := accounts[0]
	want := Account{
		Email: "me@example.com", DisplayName: "Me", Username: "login-name",
		IMAPHost: "imap.example.com", IMAPPort: 993,
		SMTPHost: "smtp.example.com", SMTPPort: 587, Kind: "imap",
	}
	if first != want {
		t.Fatalf("account = %+v, want %+v", first, want)
	}

	second := accounts[1]
	if second.Username != "" {
		t.Fatalf("username = %q, want empty when it matches the address", second.Username)
	}
	if second.SMTPHost != "smtp.example.com" || second.SMTPPort != 587 {
		t.Fatalf("default smtp server was not used: %+v", second)
	}
}

func TestLocalFoldersDirPrefersRelativePref(t *testing.T) {
	profile := t.TempDir()
	p := prefs{
		"mail.server.server2.type":          "none",
		"mail.server.server2.directory-rel": "[ProfD]Mail/Local Folders",
	}
	got := localFoldersDir(profile, p)
	want := filepath.Join(profile, "Mail", "Local Folders")
	if got != want {
		t.Fatalf("dir = %q, want %q", got, want)
	}
}

// the mail directory holds one file per folder with no extension, index
// sidecars next to them, and nested folders inside .sbd directories.
func TestScanMailDirSkipsSidecarsAndRecurses(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("Inbox", "From a@b Mon Jan  1 00:00:00 2020\n\nhi\n")
	write("Inbox.msf", "index, not mail")
	write("Archive", "From a@b Mon Jan  1 00:00:00 2020\n\nhi\n")
	write("Archive.sbd/2019", "From a@b Mon Jan  1 00:00:00 2020\n\nhi\n")
	write("Empty", "")

	files := scanMailDir(dir, "")
	var names []string
	for _, f := range files {
		names = append(names, f.Name)
	}
	want := []string{"Archive", "Archive/2019", "Inbox"}
	if len(names) != len(want) {
		t.Fatalf("folders = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("folders = %v, want %v", names, want)
		}
	}
}
