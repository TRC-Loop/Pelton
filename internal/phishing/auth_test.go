package phishing

import "testing"

func TestParseAuth(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		want    Auth
	}{
		{
			name:    "everything passes",
			headers: []string{"mx.example.com; spf=pass smtp.mailfrom=example.com; dkim=pass header.d=example.com; dmarc=pass header.from=example.com"},
			want:    Auth{SPF: "pass", SPFDomain: "example.com", DKIM: "pass", DKIMDomain: "example.com", DMARC: "pass"},
		},
		{
			name:    "results with comments",
			headers: []string{"mx.google.com; dkim=pass (test mode) header.d=example.com; spf=softfail (google.com: domain of transitioning) smtp.mailfrom=bounce.example.net"},
			want:    Auth{SPF: "softfail", SPFDomain: "bounce.example.net", DKIM: "pass", DKIMDomain: "example.com"},
		},
		{
			name:    "quoted property",
			headers: []string{`mx.example.com; dkim=fail header.d="example.com"`},
			want:    Auth{DKIM: "fail", DKIMDomain: "example.com"},
		},
		{
			name:    "header.i falls back for the signing domain",
			headers: []string{"mx.example.com; dkim=pass header.i=@mail.example.com"},
			want:    Auth{DKIM: "pass", DKIMDomain: "mail.example.com"},
		},
		{
			name: "a later hop only fills gaps",
			headers: []string{
				"mine.example.org; spf=fail smtp.mailfrom=evil.example",
				"relay.example.net; spf=pass smtp.mailfrom=evil.example; dmarc=pass",
			},
			want: Auth{SPF: "fail", SPFDomain: "evil.example", DMARC: "pass"},
		},
		{
			name:    "one passing signature among several wins",
			headers: []string{"mx.example.com; dkim=fail header.d=other.example; dkim=pass header.d=example.com"},
			want:    Auth{DKIM: "pass", DKIMDomain: "example.com"},
		},
		{
			name:    "no header at all",
			headers: nil,
			want:    Auth{},
		},
		{
			name:    "authserv-id only",
			headers: []string{"mx.example.com; none"},
			want:    Auth{},
		},
		{
			name:    "junk does not panic",
			headers: []string{";;;=;=spf", "  "},
			want:    Auth{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseAuth(tt.headers); got != tt.want {
				t.Errorf("ParseAuth() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAuthStated(t *testing.T) {
	if (Auth{}).Stated() {
		t.Error("an empty Auth claims the server said something")
	}
	if !(Auth{SPF: "none"}).Stated() {
		t.Error("spf=none is still the server having spoken")
	}
}
