package mailview

import (
	"strings"

	htmltok "golang.org/x/net/html"
)

// blockElements are the elements that end a line of prose. A quote built from
// html has to keep the sender's line structure: collapsing a message to one
// paragraph, the way PlainText does for snippets and the search index, makes an
// unreadable reply.
var blockElements = map[string]bool{
	"address": true, "article": true, "aside": true, "blockquote": true,
	"div": true, "dd": true, "dl": true, "dt": true, "fieldset": true,
	"figcaption": true, "figure": true, "footer": true, "form": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"header": true, "hr": true, "li": true, "main": true, "nav": true,
	"ol": true, "p": true, "pre": true, "section": true, "table": true,
	"tbody": true, "td": true, "tfoot": true, "th": true, "thead": true,
	"tr": true, "ul": true,
}

// TextForQuote renders html as the plain text a reply or forward quotes. Block
// elements and <br> become line breaks, entities are decoded, and script/style
// content is dropped.
//
// It is separate from PlainText, which flattens everything to a single line
// because a snippet and a search document have no use for line structure. A
// quote does: it is the sender's message, and the reader is about to write
// between its lines.
func TextForQuote(markup string) string {
	if markup == "" {
		return ""
	}
	z := htmltok.NewTokenizer(strings.NewReader(markup))
	var lines []string
	var line strings.Builder
	// depth of nested elements whose text is being discarded.
	skip := 0

	endLine := func() {
		lines = append(lines, strings.TrimRight(line.String(), " "))
		line.Reset()
	}

	for {
		switch z.Next() {
		case htmltok.ErrorToken:
			// io.EOF or malformed markup: keep whatever was recovered.
			endLine()
			return joinQuoteLines(lines)
		case htmltok.TextToken:
			if skip > 0 {
				continue
			}
			// runs of whitespace inside a line collapse, since html treats a
			// newline in the source as a space. The line breaks come from the
			// tags, not from how the sender's html happens to be wrapped.
			text := strings.Join(strings.Fields(string(z.Text())), " ")
			if text == "" {
				continue
			}
			if line.Len() > 0 && !strings.HasSuffix(line.String(), " ") {
				line.WriteByte(' ')
			}
			line.WriteString(text)
		case htmltok.StartTagToken, htmltok.SelfClosingTagToken:
			name, _ := z.TagName()
			tag := string(name)
			if skipTextIn[tag] {
				skip++
				continue
			}
			if tag == "br" || blockElements[tag] {
				endLine()
			}
		case htmltok.EndTagToken:
			name, _ := z.TagName()
			tag := string(name)
			if skipTextIn[tag] {
				if skip > 0 {
					skip--
				}
				continue
			}
			if blockElements[tag] {
				endLine()
			}
		}
	}
}

// joinQuoteLines drops the blank lines the walk produces. Adjacent blocks each
// end a line, so </p><p> yields an empty one between them, and mail wrapped in
// nested layout tables yields many.
//
// Every blank line goes, not just the runs: a quote is prefixed with "> " on
// each line, which separates paragraphs well enough on its own, and keeping
// them would mean deciding which of a layout table's empty rows the sender
// meant. The cost is that a deliberate blank line in the original is lost.
func joinQuoteLines(lines []string) string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		if l != "" {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}
