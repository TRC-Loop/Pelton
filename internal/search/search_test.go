package search

import (
	"testing"
	"time"
)

// testIndex builds a throwaway index holding docs.
func testIndex(t *testing.T, docs ...Doc) *Index {
	t.Helper()
	idx, err := Open(t.TempDir() + "/test.bleve")
	if err != nil {
		t.Fatalf("open index: %v", err)
	}
	t.Cleanup(func() { _ = idx.Close() })
	for _, d := range docs {
		if d.Date.IsZero() {
			d.Date = time.Now()
		}
		if err := idx.IndexDoc(d); err != nil {
			t.Fatalf("index doc %d: %v", d.ID, err)
		}
	}
	return idx
}

// ids runs a query and returns the matching ids in rank order.
func ids(t *testing.T, idx *Index, q Query) []int64 {
	t.Helper()
	res, err := idx.Search(q)
	if err != nil {
		t.Fatalf("search %q: %v", q.Text, err)
	}
	out := make([]int64, 0, len(res.Hits))
	for _, h := range res.Hits {
		out = append(out, h.ID)
	}
	return out
}

func contains(ids []int64, want int64) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

var corpus = []Doc{
	{ID: 1, Subject: "Invoice for March", Body: "the invoice is attached"},
	{ID: 2, Subject: "Invoices for Q1", From: "Billing billing@example.test"},
	{ID: 3, Subject: "Lunch on Friday", Body: "the place near the park"},
	{ID: 4, Subject: "Grüne Rechnung", Body: "die Rechnung ist fällig"},
}

// A word that is verbatim in a message must find it. This is the whole promise
// of the feature and the thing users reported broken.
func TestExactWordFindsItsMessage(t *testing.T) {
	idx := testIndex(t, corpus...)
	for _, tt := range []struct {
		query string
		want  int64
	}{
		{"invoice", 1},
		{"INVOICE", 1},
		{"Friday", 3},
		{"Rechnung", 4},
		{"billing", 2},
	} {
		if got := ids(t, idx, Query{Text: tt.query}); !contains(got, tt.want) {
			t.Errorf("search %q did not find message %d, got %v", tt.query, tt.want, got)
		}
	}
}

// Typing a partial word should already show results rather than looking broken
// until the last keystroke.
func TestPrefixMatchesWhileTyping(t *testing.T) {
	idx := testIndex(t, corpus...)
	got := ids(t, idx, Query{Text: "invoi"})
	if !contains(got, 1) {
		t.Errorf("prefix %q did not find message 1, got %v", "invoi", got)
	}
}

// Prefix matching is what makes a singular reach a plural without a stemmer.
func TestSingularReachesPlural(t *testing.T) {
	idx := testIndex(t, corpus...)
	got := ids(t, idx, Query{Text: "invoice"})
	if !contains(got, 2) {
		t.Errorf("search %q did not reach the plural subject, got %v", "invoice", got)
	}
}

// A term too short to fuzz safely must not drag in unrelated mail. "lun" is a
// prefix of Lunch and one edit from several other words.
func TestShortTermIsNotFuzzed(t *testing.T) {
	if got := shouldFuzz("lun"); got {
		t.Errorf("shouldFuzz(%q) = true, want false", "lun")
	}
	if got := shouldFuzz("invoice march"); !got {
		t.Errorf("shouldFuzz(%q) = false, want true", "invoice march")
	}
	// one short term is enough to disable fuzzing for the whole query.
	if got := shouldFuzz("invoice of"); got {
		t.Errorf("shouldFuzz(%q) = true, want false", "invoice of")
	}
}

// A typo within one edit still finds the mail, which is what fuzziness is for.
func TestTypoStillFinds(t *testing.T) {
	idx := testIndex(t, corpus...)
	got := ids(t, idx, Query{Text: "invoive"})
	if !contains(got, 1) {
		t.Errorf("typo %q did not find message 1, got %v", "invoive", got)
	}
}

// Total must report every match, not just the page, or the caller cannot tell a
// full page from a truncated one.
func TestTotalCountsBeyondThePage(t *testing.T) {
	docs := make([]Doc, 0, 30)
	for i := 1; i <= 30; i++ {
		docs = append(docs, Doc{ID: int64(i), Subject: "weekly report", Date: time.Now()})
	}
	idx := testIndex(t, docs...)

	res, err := idx.Search(Query{Text: "weekly", Limit: 10})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(res.Hits) != 10 {
		t.Errorf("page holds %d hits, want 10", len(res.Hits))
	}
	if res.Total != 30 {
		t.Errorf("Total = %d, want 30", res.Total)
	}
}

// Paging must reach hits ranked past the first page, and never repeat one.
func TestOffsetPagesThroughResults(t *testing.T) {
	docs := make([]Doc, 0, 30)
	for i := 1; i <= 30; i++ {
		docs = append(docs, Doc{ID: int64(i), Subject: "weekly report", Date: time.Now()})
	}
	idx := testIndex(t, docs...)

	seen := map[int64]bool{}
	for offset := 0; offset < 30; offset += 10 {
		for _, id := range ids(t, idx, Query{Text: "weekly", Limit: 10, Offset: offset}) {
			if seen[id] {
				t.Errorf("message %d returned on more than one page", id)
			}
			seen[id] = true
		}
	}
	if len(seen) != 30 {
		t.Errorf("paging reached %d of 30 messages", len(seen))
	}
}

// Field chips stay precise: no fuzziness, and every token must be present.
func TestFieldChipIsExact(t *testing.T) {
	idx := testIndex(t, corpus...)
	if got := ids(t, idx, Query{From: "billing@example.test"}); !contains(got, 2) {
		t.Errorf("from chip did not find message 2, got %v", got)
	}
	if got := ids(t, idx, Query{Subject: "Lunch"}); !contains(got, 3) {
		t.Errorf("subject chip did not find message 3, got %v", got)
	}
}
