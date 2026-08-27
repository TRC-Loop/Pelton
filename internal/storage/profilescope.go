package storage

import (
	"context"
	"sync/atomic"
)

// The active profile's scope (#270).
//
// Settings, signatures and saved views belong to a profile, and every query for
// them needs to know which. Threading a profile id through every caller would
// touch most of the app for a value that is the same everywhere: one process is
// in one profile at a time. So the store holds it, resolved once when the app
// starts or switches, and the query helpers read it.
//
// The three are separate ids rather than one, because sharing is per area. A
// profile that shares settings with main but owns its signatures resolves to
// main for the first and to itself for the second.
//
// They are atomics because a switch happens on the ui goroutine while sync
// goroutines are still reading settings.

// UseProfile points the store at a profile, resolving each area to the profile
// whose rows it actually reads: itself, or main when that area is shared.
func (d *DB) UseProfile(p Profile, mainID int64) {
	d.scope.profile.Store(p.ID)
	d.scope.settings.Store(owner(p.ID, mainID, p.ShareSettings))
	d.scope.signatures.Store(owner(p.ID, mainID, p.ShareSignatures))
	d.scope.views.Store(owner(p.ID, mainID, p.ShareViews))
	d.scope.layout.Store(owner(p.ID, mainID, p.ShareLayout))
}

// UseActiveProfile points the store at whichever profile is marked active. It
// runs at the end of RunMigrations so any opened store has a usable scope,
// tests and one-off tools included.
func (d *DB) UseActiveProfile(ctx context.Context) error {
	active, err := d.ActiveProfile(ctx)
	if err != nil {
		return err
	}
	main, err := d.MainProfile(ctx)
	if err != nil {
		return err
	}
	d.UseProfile(*active, main.ID)
	return nil
}

// ScopedProfileID reports the profile the store is currently scoped to, which
// is the active one. Areas it shares resolve elsewhere; this is the identity.
func (d *DB) ScopedProfileID() int64 {
	return orMain(d.scope.profile.Load())
}

// settingsProfile, signaturesProfile and viewsProfile are the owning profile for
// each area. A zero value means the scope was never set, which only happens
// before migrations have run; falling back to 1 keeps a half-initialised store
// reading the main profile rather than nothing at all.
func (d *DB) settingsProfile() int64 { return orMain(d.scope.settings.Load()) }

func (d *DB) signaturesProfile() int64 { return orMain(d.scope.signatures.Load()) }

func (d *DB) viewsProfile() int64 { return orMain(d.scope.views.Load()) }

// layoutProfile owns the sidebar layout rows: which folders are pinned and the
// order of folders and account sections.
func (d *DB) layoutProfile() int64 { return orMain(d.scope.layout.Load()) }

// owner picks whose rows an area reads.
func owner(profileID, mainID int64, shared bool) int64 {
	if shared {
		return mainID
	}
	return profileID
}

func orMain(id int64) int64 {
	if id == 0 {
		return firstProfileID
	}
	return id
}

// firstProfileID is the main profile's id on every install: the profiles table
// is created empty by the migration, which then inserts main as its first row.
const firstProfileID = 1

// scope holds the resolved ids. Kept in one struct so the DB fields stay
// readable.
type scope struct {
	profile    atomic.Int64
	settings   atomic.Int64
	signatures atomic.Int64
	views      atomic.Int64
	layout     atomic.Int64
}
