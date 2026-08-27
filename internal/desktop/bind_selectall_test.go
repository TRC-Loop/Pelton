package desktop

import "testing"

// a selection is an id list all the way to the bulk actions, so select-all has
// a ceiling. What matters is that it says when it hit it rather than quietly
// selecting less than it claims.
func TestCapIDsReportsWhenItStoppedShort(t *testing.T) {
	small := make([]int64, 10)
	got := capIDs(small)
	if got.Capped || len(got.IDs) != 10 || got.Matching != 10 {
		t.Errorf("capIDs(10) = %+v, want all ten and not capped", got)
	}

	big := make([]int64, selectAllCap+5)
	got = capIDs(big)
	if !got.Capped {
		t.Error("capIDs did not report a selection it cut short")
	}
	if len(got.IDs) != selectAllCap {
		t.Errorf("len(IDs) = %d, want the cap of %d", len(got.IDs), selectAllCap)
	}
	if got.Matching != selectAllCap+5 {
		t.Errorf("Matching = %d, want the untrimmed %d", got.Matching, selectAllCap+5)
	}

	// nil has to cross the bridge as an empty array, not null: the frontend
	// spreads it into a Set.
	if ids := capIDs(nil).IDs; ids == nil || len(ids) != 0 {
		t.Errorf("capIDs(nil).IDs = %v, want an empty slice", ids)
	}
}
