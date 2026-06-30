package desktop

// pins libs as direct deps until feature code imports them

import (
	_ "github.com/emersion/go-imap/v2"
	_ "github.com/emersion/go-message"
	_ "github.com/emersion/go-smtp"

	_ "golang.org/x/oauth2"

	_ "modernc.org/sqlite"

	_ "github.com/JohannesKaufmann/html-to-markdown/v2"
	_ "github.com/yuin/goldmark"

	_ "github.com/microcosm-cc/bluemonday"

	_ "golang.org/x/text/encoding"

	_ "github.com/VirusTotal/vt-go"
)
