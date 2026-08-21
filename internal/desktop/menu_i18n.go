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
	closeWindow      string
	mailboxMenu      string
	syncNow          string
	openWindow       string
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
	openInTab        string
	switchProfile    string
	closeTab         string
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
		closeWindow:      "Close Window",
		mailboxMenu:      "Mailbox",
		syncNow:          "Sync Now",
		openWindow:       "Open Pelton",
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
		openInTab:        "Open in New Tab",
		switchProfile:    "Switch Profile…",
		closeTab:         "Close Tab",
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
		closeWindow:      "Fenster schließen",
		mailboxMenu:      "Postfach",
		syncNow:          "Jetzt synchronisieren",
		openWindow:       "Pelton öffnen",
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
		openInTab:        "In neuem Tab öffnen",
		switchProfile:    "Profil wechseln…",
		closeTab:         "Tab schließen",
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
		closeWindow:      "Fermer la fenêtre",
		mailboxMenu:      "Boîte mail",
		syncNow:          "Synchroniser maintenant",
		openWindow:       "Ouvrir Pelton",
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
		openInTab:        "Ouvrir dans un nouvel onglet",
		switchProfile:    "Changer de profil…",
		closeTab:         "Fermer l'onglet",
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
		closeWindow:      "Venster sluiten",
		mailboxMenu:      "Mailbox",
		syncNow:          "Nu synchroniseren",
		openWindow:       "Pelton openen",
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
		openInTab:        "Openen in nieuw tabblad",
		switchProfile:    "Profiel wisselen…",
		closeTab:         "Tabblad sluiten",
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
		closeWindow:      "Cerrar ventana",
		mailboxMenu:      "Buzón",
		syncNow:          "Sincronizar ahora",
		openWindow:       "Abrir Pelton",
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
		openInTab:        "Abrir en una pestaña nueva",
		switchProfile:    "Cambiar de perfil…",
		closeTab:         "Cerrar pestaña",
	},
	"pl": {
		appMenu:          "Pelton",
		about:            "O programie Pelton",
		preferences:      "Preferencje…",
		hide:             "Ukryj Pelton",
		quit:             "Zakończ Pelton",
		fileMenu:         "Plik",
		compose:          "Napisz",
		exportPDF:        "Eksportuj wiadomość jako PDF…",
		closeWindow:      "Zamknij okno",
		mailboxMenu:      "Skrzynka",
		syncNow:          "Synchronizuj teraz",
		openWindow:       "Otwórz Pelton",
		addMailbox:       "Dodaj skrzynkę…",
		manageMailboxes:  "Zarządzaj skrzynkami…",
		mailMenu:         "Wiadomość",
		undo:             "Cofnij",
		markRead:         "Oznacz jako przeczytane",
		markUnread:       "Oznacz jako nieprzeczytane",
		flagUnflag:       "Oznacz / usuń flagę",
		archive:          "Archiwizuj",
		deleteMessage:    "Usuń wiadomość",
		viewMenu:         "Widok",
		toggleFullscreen: "Przełącz pełny ekran",
		lowPowerMode:     "Tryb oszczędzania energii",
		openInTab:        "Otwórz w nowej karcie",
		switchProfile:    "Zmień profil…",
		closeTab:         "Zamknij kartę",
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
