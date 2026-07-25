package desktop

import (
	"strings"

	"github.com/TRC-Loop/Pelton/internal/storage"
	"github.com/gen2brain/beeep"
)

// Native OS notifications for new mail (#126). beeep raises a real notification
// on macOS (Notification Center), Windows (toast) and Linux (dbus). It has no
// per-notification click callback, so a click does not open a specific message;
// exact click-to-open would need platform-specific code and is tracked
// separately. Notifications only fire for INBOX so sent/archived mail moved by
// a sync never notifies.

// notifyStrings holds the localizable text of a new-mail notification. Like the
// native menu (see menu_i18n.go), notifications are built in the Go process, so
// their text lives in a Go-side table keyed by the language setting rather than
// the frontend's locale files.
type notifyStrings struct {
	newMail     string // title for a plain new-mail notification
	vipNewMail  string // title when the sender is a VIP
	noSubject   string // shown in place of an empty subject
	fromUnknown string // shown in place of an empty sender
}

var notifyLocales = map[string]notifyStrings{
	"en": {newMail: "New mail", vipNewMail: "New mail from a VIP", noSubject: "(no subject)", fromUnknown: "Unknown sender"},
	"de": {newMail: "Neue Nachricht", vipNewMail: "Neue Nachricht von einem VIP", noSubject: "(kein Betreff)", fromUnknown: "Unbekannter Absender"},
	"fr": {newMail: "Nouveau message", vipNewMail: "Nouveau message d'un VIP", noSubject: "(sans objet)", fromUnknown: "Expéditeur inconnu"},
	"nl": {newMail: "Nieuw bericht", vipNewMail: "Nieuw bericht van een VIP", noSubject: "(geen onderwerp)", fromUnknown: "Onbekende afzender"},
	"es": {newMail: "Nuevo mensaje", vipNewMail: "Nuevo mensaje de un VIP", noSubject: "(sin asunto)", fromUnknown: "Remitente desconocido"},
}

// notifyStringsFor returns the notification text table for lang, falling back to
// English for an unknown or empty code.
func notifyStringsFor(lang string) notifyStrings {
	if s, ok := notifyLocales[lang]; ok {
		return s
	}
	return notifyLocales["en"]
}

// notifyNewMail raises OS notifications for messages freshly pulled into a
// folder. It runs on its own goroutine off the sync path. General new-mail
// notifications only fire when the setting is on; VIP-sender notifications fire
// regardless, so important senders cut through a quiet inbox. Only INBOX is
// considered so mail landing in other folders during a full sync stays quiet.
func (a *App) notifyNewMail(folder storage.Folder, ids []int64) {
	if len(ids) == 0 || !strings.EqualFold(folder.IMAPPath, "INBOX") {
		return
	}
	general := a.boolSetting(settingNotifyNewMail, false)
	vips := a.vipSet()
	if !general && len(vips) == 0 {
		return
	}

	s := notifyStringsFor(a.stringSetting(settingLanguage, "en"))
	for _, id := range ids {
		if a.ctx.Err() != nil {
			return
		}
		m, err := a.store.GetMessage(a.ctx, id)
		if err != nil {
			continue
		}
		// never notify for mail the user has already seen (e.g. read elsewhere and
		// pulled in on a first sync).
		if m.Flags.Has(storage.FlagSeen) {
			continue
		}
		vip := vips[bareAddress(m.FromAddress)]
		if !vip && !general {
			continue
		}

		title := s.newMail
		if vip {
			title = s.vipNewMail
		}
		a.sendNotification(title, notifyBody(m, s))
	}
}

// notifyBody builds the notification body: sender then subject, each falling
// back to a placeholder so the notification is never blank.
func notifyBody(m *storage.Message, s notifyStrings) string {
	from := strings.TrimSpace(m.FromName)
	if from == "" {
		from = bareAddress(m.FromAddress)
	}
	if from == "" {
		from = s.fromUnknown
	}
	subject := strings.TrimSpace(m.Subject)
	if subject == "" {
		subject = s.noSubject
	}
	return from + "\n" + subject
}

// sendNotification raises one native OS notification, logging (not failing) on
// error so a notification backend that is missing or busy never disrupts sync.
func (a *App) sendNotification(title, body string) {
	if err := beeep.Notify(title, body, ""); err != nil {
		a.log.Warn("notify", "err", err)
	}
}
