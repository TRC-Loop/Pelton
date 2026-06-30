package desktop

import (
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// signatures: reusable header/footer blocks the user manages in settings, assigns
// as per-mailbox defaults, and inserts (or swaps) in compose. The store holds the
// blocks and the per-account assignment; these bindings are the typed boundary.

// SignatureDTO mirrors storage.Signature for the frontend. Kind is "header" or
// "footer"; Format is "markdown" or "html".
type SignatureDTO struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Format  string `json:"format"`
	Content string `json:"content"`
}

// AccountSignaturesDTO is an account's default header/footer ids; 0 means none.
type AccountSignaturesDTO struct {
	HeaderID int64 `json:"headerId"`
	FooterID int64 `json:"footerId"`
}

func toSignatureDTO(s storage.Signature) SignatureDTO {
	return SignatureDTO{ID: s.ID, Name: s.Name, Kind: s.Kind, Format: s.Format, Content: s.Content}
}

// ListSignatures returns every signature block.
func (a *App) ListSignatures() ([]SignatureDTO, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	sigs, err := a.store.ListSignatures(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]SignatureDTO, 0, len(sigs))
	for _, s := range sigs {
		out = append(out, toSignatureDTO(s))
	}
	return out, nil
}

// SaveSignature creates the signature when ID is 0, otherwise updates it, and
// returns the stored row (with its id).
func (a *App) SaveSignature(dto SignatureDTO) (SignatureDTO, error) {
	if err := a.ready(); err != nil {
		return SignatureDTO{}, err
	}
	s := storage.Signature{
		ID:      dto.ID,
		Name:    dto.Name,
		Kind:    dto.Kind,
		Format:  dto.Format,
		Content: dto.Content,
	}
	if s.ID == 0 {
		if _, err := a.store.CreateSignature(a.ctx, &s); err != nil {
			return SignatureDTO{}, err
		}
	} else if err := a.store.UpdateSignature(a.ctx, &s); err != nil {
		return SignatureDTO{}, err
	}
	return toSignatureDTO(s), nil
}

// DeleteSignature removes a signature block. Any account that defaulted to it has
// that slot cleared by the database.
func (a *App) DeleteSignature(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.DeleteSignature(a.ctx, id)
}

// GetAccountSignatures returns an account's default header/footer ids.
func (a *App) GetAccountSignatures(accountID int64) (AccountSignaturesDTO, error) {
	if err := a.ready(); err != nil {
		return AccountSignaturesDTO{}, err
	}
	as, err := a.store.GetAccountSignatures(a.ctx, accountID)
	if err != nil {
		return AccountSignaturesDTO{}, err
	}
	return AccountSignaturesDTO{HeaderID: as.HeaderID, FooterID: as.FooterID}, nil
}

// SetAccountSignatures sets an account's default header/footer assignment; pass 0
// to clear a slot.
func (a *App) SetAccountSignatures(accountID, headerID, footerID int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.SetAccountSignatures(a.ctx, accountID, storage.AccountSignatures{
		HeaderID: headerID,
		FooterID: footerID,
	})
}
