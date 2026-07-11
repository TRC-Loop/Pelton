package desktop

// menuStrings holds every literal label buildMenu uses. Menu items that carry
// a native "%s Pelton"-style app name (Hide/Quit/About) keep the app name as a
// separate substitution since it's never translated.
type menuStrings struct {
	appMenu          string
	about            string
	preferences      string
	hide             string
	quit             string
	fileMenu         string
	compose          string
	exportPDF        string
	mailboxMenu      string
	syncNow          string
	addMailbox       string
	manageMailboxes  string
	mailMenu         string
	undo             string
	markRead         string
	markUnread       string
	flagUnflag       string
	archive          string
	deleteMessage    string
	viewMenu         string
	toggleFullscreen string
	lowPowerMode     string
}

// menuLocales mirrors the frontend's supported languages (en/de/fr/nl/es).
// Keeping this as a Go-side table, separate from the frontend's TypeScript
// locale files, is unavoidable: the native menu is built by the Go process
// before any frontend code runs, and wails' menu API takes plain strings, not
// translation keys the webview could resolve later.
var menuLocales = map[string]menuStrings{
	"en": {
		appMenu:          "Pelton",
		about:            "About Pelton",
		preferences:      "Preferences…",
		hide:             "Hide Pelton",
		quit:             "Quit Pelton",
		fileMenu:         "File",
		compose:          "Compose",
		exportPDF:        "Export Message as PDF…",
		mailboxMenu:      "Mailbox",
		syncNow:          "Sync Now",
		addMailbox:       "Add Mailbox…",
		manageMailboxes:  "Manage Mailboxes…",
		mailMenu:         "Mail",
		undo:             "Undo",
		markRead:         "Mark as Read",
		markUnread:       "Mark as Unread",
		flagUnflag:       "Flag / Unflag",
		archive:          "Archive",
		deleteMessage:    "Delete Message",
		viewMenu:         "View",
		toggleFullscreen: "Toggle Fullscreen",
		lowPowerMode:     "Low Power Mode",
	},
	"de": {
		appMenu:          "Pelton",
		about:            "Über Pelton",
		preferences:      "Einstellungen…",
		hide:             "Pelton ausblenden",
		quit:             "Pelton beenden",
		fileMenu:         "Datei",
		compose:          "Neue Nachricht",
		exportPDF:        "Nachricht als PDF exportieren…",
		mailboxMenu:      "Postfach",
		syncNow:          "Jetzt synchronisieren",
		addMailbox:       "Postfach hinzufügen…",
		manageMailboxes:  "Postfächer verwalten…",
		mailMenu:         "Nachricht",
		undo:             "Rückgängig",
		markRead:         "Als gelesen markieren",
		markUnread:       "Als ungelesen markieren",
		flagUnflag:       "Markieren / Demarkieren",
		archive:          "Archivieren",
		deleteMessage:    "Nachricht löschen",
		viewMenu:         "Ansicht",
		toggleFullscreen: "Vollbild umschalten",
		lowPowerMode:     "Energiesparmodus",
	},
	"fr": {
		appMenu:          "Pelton",
		about:            "À propos de Pelton",
		preferences:      "Préférences…",
		hide:             "Masquer Pelton",
		quit:             "Quitter Pelton",
		fileMenu:         "Fichier",
		compose:          "Nouveau message",
		exportPDF:        "Exporter le message en PDF…",
		mailboxMenu:      "Boîte mail",
		syncNow:          "Synchroniser maintenant",
		addMailbox:       "Ajouter une boîte mail…",
		manageMailboxes:  "Gérer les boîtes mail…",
		mailMenu:         "Message",
		undo:             "Annuler",
		markRead:         "Marquer comme lu",
		markUnread:       "Marquer comme non lu",
		flagUnflag:       "Marquer / Démarquer",
		archive:          "Archiver",
		deleteMessage:    "Supprimer le message",
		viewMenu:         "Affichage",
		toggleFullscreen: "Basculer le plein écran",
		lowPowerMode:     "Mode basse consommation",
	},
	"nl": {
		appMenu:          "Pelton",
		about:            "Over Pelton",
		preferences:      "Voorkeuren…",
		hide:             "Pelton verbergen",
		quit:             "Pelton afsluiten",
		fileMenu:         "Bestand",
		compose:          "Nieuw bericht",
		exportPDF:        "Bericht exporteren als PDF…",
		mailboxMenu:      "Mailbox",
		syncNow:          "Nu synchroniseren",
		addMailbox:       "Mailbox toevoegen…",
		manageMailboxes:  "Mailboxen beheren…",
		mailMenu:         "Bericht",
		undo:             "Ongedaan maken",
		markRead:         "Als gelezen markeren",
		markUnread:       "Als ongelezen markeren",
		flagUnflag:       "Markeren / demarkeren",
		archive:          "Archiveren",
		deleteMessage:    "Bericht verwijderen",
		viewMenu:         "Beeld",
		toggleFullscreen: "Volledig scherm in-/uitschakelen",
		lowPowerMode:     "Energiebesparende modus",
	},
	"es": {
		appMenu:          "Pelton",
		about:            "Acerca de Pelton",
		preferences:      "Preferencias…",
		hide:             "Ocultar Pelton",
		quit:             "Salir de Pelton",
		fileMenu:         "Archivo",
		compose:          "Redactar",
		exportPDF:        "Exportar mensaje como PDF…",
		mailboxMenu:      "Buzón",
		syncNow:          "Sincronizar ahora",
		addMailbox:       "Añadir buzón…",
		manageMailboxes:  "Administrar buzones…",
		mailMenu:         "Mensaje",
		undo:             "Deshacer",
		markRead:         "Marcar como leído",
		markUnread:       "Marcar como no leído",
		flagUnflag:       "Marcar / Desmarcar",
		archive:          "Archivar",
		deleteMessage:    "Eliminar mensaje",
		viewMenu:         "Ver",
		toggleFullscreen: "Alternar pantalla completa",
		lowPowerMode:     "Modo de bajo consumo",
	},
}

// menuStringsFor returns the translation table for lang, falling back to
// English for an unrecognized or empty language code.
func menuStringsFor(lang string) menuStrings {
	if s, ok := menuLocales[lang]; ok {
		return s
	}
	return menuLocales["en"]
}
