package desktop

import (
	"strings"
	"time"

	"github.com/TRC-Loop/Pelton/internal/mailview"
	"github.com/TRC-Loop/Pelton/internal/storage"
)

// The dtos below are the flat shapes the frontend receives. They exist so the ui
// contract is stable and json friendly and does not leak internal struct details
// like bitmask flags or go time values. Wails generates typescript interfaces
// from these, which src/lib/api.ts re-exports.

// AccountDTO is non sensitive account metadata for the sidebar.
type AccountDTO struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	IMAPHost    string `json:"imapHost"`
	IMAPPort    int    `json:"imapPort"`
	SMTPHost    string `json:"smtpHost"`
	SMTPPort    int    `json:"smtpPort"`
}

// FolderDTO is one mailbox in an account's tree. ParentID is null at the root.
// Role classifies known special folders so the ui can icon them and so unified
// views can gather them. UnreadCount/TotalCount drive the badges.
type FolderDTO struct {
	ID          int64    `json:"id"`
	AccountID   int64    `json:"accountId"`
	Name        string   `json:"name"`
	IMAPPath    string   `json:"imapPath"`
	Delimiter   string   `json:"delimiter"`
	ParentID    *int64   `json:"parentId"`
	Role        string   `json:"role"`
	UnreadCount int      `json:"unreadCount"`
	TotalCount  int      `json:"totalCount"`
	Attributes  []string `json:"attributes"`
}

// UnifiedViewDTO is a cross account view (inbox/flagged/sent/drafts). The ui
// passes Key back to ListMessages to read the merged list.
type UnifiedViewDTO struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	UnreadCount int    `json:"unreadCount"`
	TotalCount  int    `json:"totalCount"`
}

// MessageSummaryDTO is one row in the message list. It carries the badge data
// (account email, folder name) so the list is self contained for unified views.
// Auth is always "unavailable" for now: Authentication-Results parsing is a
// documented backend follow-up (see bind_messages.go).
type MessageSummaryDTO struct {
	ID             int64  `json:"id"`
	AccountID      int64  `json:"accountId"`
	FolderID       int64  `json:"folderId"`
	AccountEmail   string `json:"accountEmail"`
	FolderName     string `json:"folderName"`
	Subject        string `json:"subject"`
	FromName       string `json:"fromName"`
	FromAddress    string `json:"fromAddress"`
	Snippet        string `json:"snippet"`
	Date           string `json:"date"`
	Seen           bool   `json:"seen"`
	Flagged        bool   `json:"flagged"`
	HasAttachments bool   `json:"hasAttachments"`
	PGP            string `json:"pgp"`
	Auth           string `json:"auth"`
	// FlagColor is 0 (none) or 1..8. Offline marks a user-pinned message.
	// SnoozeUntil is a stored timestamp (empty when not snoozed).
	FlagColor   int    `json:"flagColor"`
	Offline     bool   `json:"offline"`
	SnoozeUntil string `json:"snoozeUntil"`
}

// MessageListDTO is a page of summaries plus the unfiltered total for paging.
type MessageListDTO struct {
	Messages []MessageSummaryDTO `json:"messages"`
	Total    int                 `json:"total"`
}

// AttachmentDTO describes a stored attachment. Inline parts (referenced by the
// html body via cid) are flagged so the ui can hide them from the download list.
type AttachmentDTO struct {
	ID          int64  `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
	Inline      bool   `json:"inline"`
}

// MessageDetailDTO is the full message for the reading pane. BodyHTMLSafe is
// already sanitized with remote content blocked and cid images inlined, ready to
// drop into the sandboxed renderer. HasRemoteContent tells the ui whether to
// offer "load remote images".
type MessageDetailDTO struct {
	MessageSummaryDTO
	ToAddresses      string `json:"toAddresses"`
	CcAddresses      string `json:"ccAddresses"`
	BodyPlain        string `json:"bodyPlain"`
	BodyHTMLSafe     string `json:"bodyHtmlSafe"`
	IsHTML           bool   `json:"isHtml"`
	HasRemoteContent bool   `json:"hasRemoteContent"`
	// RemoteAllowed is true when this message's remote content was rendered
	// because the sender or domain is trusted (or the global setting is on), so
	// the ui shows no blocked-images banner.
	RemoteAllowed bool `json:"remoteAllowed"`
	// RemoteHosts lists the hosts the blocked remote content would load from, so
	// the banner can show the user where.
	RemoteHosts []string        `json:"remoteHosts"`
	Attachments []AttachmentDTO `json:"attachments"`
	// Unsubscribe describes the unsubscribe mechanism the message advertises
	// via its List-Unsubscribe headers, nil when it has none on record.
	Unsubscribe *UnsubscribeDTO `json:"unsubscribe"`
}

// authUnavailable is the placeholder auth status. The backend does not yet parse
// Authentication-Results headers into storage, so spf/dkim/dmarc are unknown.
// TODO(backend): parse Authentication-Results during sync into a new column and
// return a structured status here instead of this placeholder.
const authUnavailable = "unavailable"

// folder roles. These match storage's known folder name constants and the
// special-use imap attributes.
const (
	roleInbox   = "inbox"
	roleSent    = "sent"
	roleDrafts  = "drafts"
	roleTrash   = "trash"
	roleJunk    = "junk"
	roleArchive = "archive"
	roleNormal  = "normal"
)

// toAccountDTO flattens a storage account.
func toAccountDTO(a storage.Account) AccountDTO {
	return AccountDTO{
		ID:          a.ID,
		Email:       a.Email,
		DisplayName: a.DisplayName,
		Username:    a.Username,
		IMAPHost:    a.IMAPHost,
		IMAPPort:    a.IMAPPort,
		SMTPHost:    a.SMTPHost,
		SMTPPort:    a.SMTPPort,
	}
}

// toFolderDTO flattens a storage folder and tags its role. Counts are filled by
// the caller since they need a second query.
func toFolderDTO(f storage.Folder) FolderDTO {
	return FolderDTO{
		ID:         f.ID,
		AccountID:  f.AccountID,
		Name:       f.Name,
		IMAPPath:   f.IMAPPath,
		Delimiter:  f.Delimiter,
		ParentID:   f.ParentID,
		Role:       folderRole(f),
		Attributes: f.Attributes,
	}
}

// folderRole classifies a folder by its special-use attributes first (the
// reliable signal) then by its name as a fallback.
func folderRole(f storage.Folder) string {
	for _, attr := range f.Attributes {
		switch strings.ToLower(strings.TrimPrefix(attr, "\\")) {
		case "sent":
			return roleSent
		case "drafts":
			return roleDrafts
		case "trash":
			return roleTrash
		case "junk":
			return roleJunk
		case "archive":
			return roleArchive
		}
	}
	switch strings.ToUpper(f.IMAPPath) {
	case storage.FolderInbox:
		return roleInbox
	case strings.ToUpper(storage.FolderSent):
		return roleSent
	case strings.ToUpper(storage.FolderDrafts):
		return roleDrafts
	case strings.ToUpper(storage.FolderTrash):
		return roleTrash
	case strings.ToUpper(storage.FolderJunk):
		return roleJunk
	case strings.ToUpper(storage.FolderArchive):
		return roleArchive
	}
	return roleNormal
}

// toSummaryDTO flattens a stored message into a list row. accountEmail and
// folderName are looked up by the caller and passed in for the badge.
func toSummaryDTO(m storage.Message, accountEmail, folderName string) MessageSummaryDTO {
	return MessageSummaryDTO{
		ID:             m.ID,
		AccountID:      m.AccountID,
		FolderID:       m.FolderID,
		AccountEmail:   accountEmail,
		FolderName:     folderName,
		Subject:        m.Subject,
		FromName:       m.FromName,
		FromAddress:    m.FromAddress,
		Snippet:        mailview.Snippet(m.BodyPlain, m.BodyHTML),
		Date:           formatDate(m.Date),
		Seen:           m.Flags.Has(storage.FlagSeen),
		Flagged:        m.Flags.Has(storage.FlagFlagged),
		HasAttachments: m.HasAttachments,
		PGP:            string(mailview.DetectPGP(m.BodyPlain, m.BodyHTML)),
		Auth:           authUnavailable,
		FlagColor:      m.FlagColor,
		Offline:        m.Offline,
		SnoozeUntil:    m.SnoozeUntil,
	}
}

// formatDate renders a message date as rfc3339 for the ui, or empty for the zero
// time so the frontend can show a neutral state.
func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
