package desktop

import "github.com/TRC-Loop/Pelton/internal/storage"

// AddressBookEntryDTO is one autocomplete/contact entry for the frontend.
type AddressBookEntryDTO struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	UseCount  int    `json:"useCount"`
	LastUsed  string `json:"lastUsed"`
	CreatedAt string `json:"createdAt"`
}

// SearchAddresses returns compose-autocomplete candidates matching query, ranked
// by how often and how recently the user has corresponded with them.
func (a *App) SearchAddresses(query string, limit int) ([]AddressBookEntryDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 8
	}
	entries, err := a.store.SearchAddresses(a.ctx, query, limit)
	if err != nil {
		return nil, err
	}
	return toAddressDTOs(entries), nil
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

// harvestAddressBook seeds the book from cached senders. It runs in the
// background at startup and after syncs so autocomplete keeps learning.
func (a *App) harvestAddressBook() {
	if a.store == nil {
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
