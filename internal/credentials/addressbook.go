package credentials

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/zalando/go-keyring"
)

// addressBookPrefix namespaces an address book's keyring entry. Per-account
// entries are the bare decimal account id, so a book's id has to be prefixed or
// book 3 and account 3 would be the same entry.
const addressBookPrefix = "addressbook-"

// StoreAddressBookPassword saves the login for a CardDAV address book, or
// clears it when empty. Like every other secret it stays out of the settings db.
func StoreAddressBookPassword(bookID int64, password string) error {
	if password == "" {
		return DeleteAddressBookPassword(bookID)
	}
	remember(password)
	if err := keyring.Set(service, addressBookKey(bookID), password); err != nil {
		return fmt.Errorf("credentials: store address book %d password: %w", bookID, err)
	}
	return nil
}

// LoadAddressBookPassword returns an address book's stored password, or "" when
// none is set. A book with no password is one that has not finished setup, not
// an error.
func LoadAddressBookPassword(bookID int64) (string, error) {
	raw, err := keyring.Get(service, addressBookKey(bookID))
	if errors.Is(err, keyring.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("credentials: load address book %d password: %w", bookID, err)
	}
	remember(raw)
	return raw, nil
}

// DeleteAddressBookPassword removes an address book's password. A missing entry
// is not an error, so removing a book that never authenticated still cleans up.
func DeleteAddressBookPassword(bookID int64) error {
	err := keyring.Delete(service, addressBookKey(bookID))
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("credentials: delete address book %d password: %w", bookID, err)
	}
	return nil
}

func addressBookKey(bookID int64) string {
	return addressBookPrefix + strconv.FormatInt(bookID, 10)
}
