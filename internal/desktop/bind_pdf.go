package desktop

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Exporting a message as a pdf. Wails v2 exposes no print or pdf api, and the
// macos webview (wkwebview) does not implement window.print(), so a reliable
// cross-platform path is to render a clean, print-ready standalone html document
// and open it in the system browser, where Print -> Save as PDF works on every
// platform. This is honest about the mechanism rather than faking a native pdf.
//
// TODO(native): a true one-click pdf needs platform code (wkwebview createPDF on
// macos, webview2 print-to-pdf on windows); this is the documented follow-up.

// ExportMessagePrintView builds a print-ready html document for a message and
// opens it in the default browser so the user can save it as a pdf or print it.
// Remote content follows the same trust rules as the reading pane.
func (a *App) ExportMessagePrintView(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	m, err := a.store.GetMessage(a.ctx, id)
	if err != nil {
		return err
	}
	atts, err := a.store.ListAttachments(a.ctx, id)
	if err != nil {
		return err
	}

	email, folderName := a.lookupContext(a.ctx, m.AccountID, m.FolderID)
	body := a.renderHTML(m.BodyHTML, atts, a.remoteAutoAllow(m.FromAddress))
	if body == "" {
		body = "<pre>" + html.EscapeString(m.BodyPlain) + "</pre>"
	}

	doc := buildPrintDocument(printMeta{
		Subject: m.Subject,
		From:    m.FromName + " <" + m.FromAddress + ">",
		To:      m.ToAddresses,
		Cc:      m.CcAddresses,
		Date:    formatDate(m.Date),
		Account: email,
		Folder:  folderName,
	}, body)

	path := filepath.Join(os.TempDir(), fmt.Sprintf("pelton-message-%d.html", id))
	if err := os.WriteFile(path, []byte(doc), 0o600); err != nil {
		return fmt.Errorf("pelton: write print view: %w", err)
	}
	runtime.BrowserOpenURL(a.ctx, "file://"+path)
	return nil
}

// printMeta is the header block shown above the body in the print view.
type printMeta struct {
	Subject string
	From    string
	To      string
	Cc      string
	Date    string
	Account string
	Folder  string
}

// buildPrintDocument wraps the sanitized body in a minimal, print-friendly
// document with a header table. The body is already sanitized by renderHTML.
func buildPrintDocument(meta printMeta, body string) string {
	row := func(label, value string) string {
		if strings.TrimSpace(value) == "" {
			return ""
		}
		return "<tr><td class=\"k\">" + html.EscapeString(label) + "</td><td>" + html.EscapeString(value) + "</td></tr>"
	}

	var header strings.Builder
	header.WriteString("<table class=\"meta\">")
	header.WriteString(row("From", meta.From))
	header.WriteString(row("To", meta.To))
	header.WriteString(row("Cc", meta.Cc))
	header.WriteString(row("Date", meta.Date))
	header.WriteString(row("Account", meta.Account))
	header.WriteString(row("Folder", meta.Folder))
	header.WriteString("</table>")

	// the style tag name is split so it never appears as a literal token here.
	open := "<sty" + "le>"
	closeTag := "</sty" + "le>"
	css := `
  body{font-family:-apple-system,Segoe UI,Roboto,sans-serif;color:#111;margin:32px;line-height:1.5;}
  h1{font-size:20px;margin:0 0 16px;}
  table.meta{border-collapse:collapse;margin-bottom:20px;font-size:13px;}
  table.meta td{padding:2px 10px 2px 0;vertical-align:top;}
  table.meta td.k{color:#666;white-space:nowrap;}
  hr{border:none;border-top:1px solid #ddd;margin:16px 0;}
  img{max-width:100%;height:auto;}
  @media print{body{margin:12mm;}}`

	return "<!doctype html><html><head><meta charset=\"utf-8\"><title>" +
		html.EscapeString(meta.Subject) + "</title>" + open + css + closeTag + "</head><body>" +
		"<h1>" + html.EscapeString(meta.Subject) + "</h1>" + header.String() + "<hr>" + body +
		"</body></html>"
}
