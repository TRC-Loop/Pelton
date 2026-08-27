package desktop

import "github.com/TRC-Loop/Pelton/internal/storage"

// AddressBookEntryDTO is one autocomplete/contact entry for the frontend.
type AddressBookEntryDTO struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	UseCount  int    `json:"useCount"`
	LastUsed  string `json:"lastUsed"`
	CreatedAt string `json:"createdAt"`
	// Contact is true for an entry from a synced address book rather than one
	// harvested from mail, so the ui can mark which is which.
	Contact bool `json:"contact"`
}

// SearchAddresses returns compose-autocomplete candidates matching query. The
// real address book comes first, then the addresses harvested from mail, ranked
// by how often and how recently the user has corresponded with them. Where the
// two hold the same address the contact's name wins: that is the one the user
// maintains, rather than whatever a sender once put in a From header (#168).
func (a *App) SearchAddresses(query string, limit int) ([]AddressBookEntryDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 8
	}
	contacts, err := a.store.SearchContacts(a.ctx, query, limit)
	if err != nil {
		return nil, err
	}
	out := make([]AddressBookEntryDTO, 0, limit)
	seen := make(map[string]bool, limit)
	for _, c := range contacts {
		if seen[c.Email] {
			continue
		}
		seen[c.Email] = true
		out = append(out, AddressBookEntryDTO{Email: c.Email, Name: c.Name, Contact: true})
	}
	if len(out) >= limit || !a.harvestAddresses() {
		return out[:min(len(out), limit)], nil
	}

	entries, err := a.store.SearchAddresses(a.ctx, query, limit)
	if err != nil {
		return nil, err
	}
	for _, e := range toAddressDTOs(entries) {
		if seen[e.Email] || len(out) >= limit {
			continue
		}
		seen[e.Email] = true
		out = append(out, e)
	}
	return out, nil
}

// ListAddresses returns the whole harvested address book for the settings
// manager, so the user can review and remove entries.
func (a *App) ListAddresses() ([]AddressBookEntryDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	entries, err := a.store.ListAddresses(a.ctx)
	if err != nil {
		return nil, err
	}
	return toAddressDTOs(entries), nil
}

// DeleteAddress removes one contact from the address book.
func (a *App) DeleteAddress(email string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.DeleteAddress(a.ctx, email)
}

// harvestAddresses reports whether addresses seen in mail are still learned.
// A user who keeps a real address book can turn it off and be offered only
// their own contacts (#168).
func (a *App) harvestAddresses() bool {
	return a.boolSetting(settingHarvestAddresses, true)
}

// harvestAddressBook seeds the book from cached senders. It runs in the
// background at startup and after syncs so autocomplete keeps learning.
func (a *App) harvestAddressBook() {
	if a.store == nil || !a.harvestAddresses() {
		return
	}
	if err := a.store.HarvestSenders(a.ctx); err != nil {
		a.log.Error("harvest address book", "err", err)
	}
}

func toAddressDTOs(entries []storage.AddressBookEntry) []AddressBookEntryDTO {
	out := make([]AddressBookEntryDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, AddressBookEntryDTO{
			Email:     e.Email,
			Name:      e.Name,
			UseCount:  e.UseCount,
			LastUsed:  e.LastUsed,
			CreatedAt: e.CreatedAt,
		})
	}
	return out
}
