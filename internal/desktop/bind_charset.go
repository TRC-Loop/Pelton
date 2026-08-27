package desktop

import (
	"github.com/TRC-Loop/Pelton/internal/charsetguess"
)

// settingCharsetFallback is how text with no usable charset is read: "auto",
// which detects it from the bytes, or the name of an encoding to use instead.
// Detection is right about ordinary mail and it is what other clients do, but
// someone who mostly receives mail from one badly configured system knows
// better than any detector, so the choice is theirs.
const settingCharsetFallback = "charset_fallback"

// settingCharsetScanned records that the one-time scan for mail cached before
// charset detection existed has run. The scan walks the whole message table, so
// it happens once rather than on every start.
const settingCharsetScanned = "charset_rescan_done"

// applyCharsetFallback pushes the setting into the parser. It is called at
// startup and whenever the setting changes, since the parser reads it on every
// message rather than being handed it.
func (a *App) applyCharsetFallback() {
	charsetguess.SetFallback(a.stringSetting(settingCharsetFallback, charsetguess.Auto))
}

// markMangledMail finds messages cached before charset detection existed whose
// stored text is not valid utf-8 and marks them to be fetched again. The raw
// source is not kept, so nothing local can decode them properly now; the next
// sync of their folder replaces the text a few at a time.
func (a *App) markMangledMail() {
	if a.boolSetting(settingCharsetScanned, false) {
		return
	}
	found, err := a.store.MarkMangledMessages(a.ctx)
	if err != nil {
		a.log.Error("scan cached mail for broken text", "err", err)
		return
	}
	if err := a.store.SetBool(a.ctx, settingCharsetScanned, true); err != nil {
		a.log.Error("record charset scan", "err", err)
	}
	if found > 0 {
		a.log.Info("marked cached mail with broken text for refetch", "count", found)
	}
}
