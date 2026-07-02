// types.ts defines the dto shapes the ui works with as plain interfaces. they
// mirror the wails-generated go dtos (frontend/wailsjs/go/models.ts) field for
// field, but are decoupled from the generated classes so the ui can spread and
// construct them freely. the generated class instances returned by the bindings
// are structurally assignable to these interfaces; api.ts wraps the few request
// types back into their generated classes before calling the bindings.

export interface Account {
  id: number
  email: string
  displayName: string
  imapHost: string
  smtpHost: string
}

export interface Folder {
  id: number
  accountId: number
  name: string
  imapPath: string
  delimiter: string
  // null/undefined at the tree root; a parent folder id otherwise. optional to
  // match the generated dto where the go nil pointer becomes undefined.
  parentId?: number | null
  role: string
  unreadCount: number
  totalCount: number
  attributes: string[]
}

export interface UnifiedView {
  key: string
  label: string
  unreadCount: number
  totalCount: number
}

export interface MessageSummary {
  id: number
  accountId: number
  folderId: number
  accountEmail: string
  folderName: string
  subject: string
  fromName: string
  fromAddress: string
  snippet: string
  date: string
  seen: boolean
  flagged: boolean
  hasAttachments: boolean
  pgp: string
  auth: string
  // flagColor is 0 (none) or 1..8. offline marks a user-pinned message.
  // snoozeUntil is a stored timestamp (empty when not snoozed).
  flagColor: number
  offline: boolean
  snoozeUntil: string
}

export interface MessageDetail extends MessageSummary {
  toAddresses: string
  ccAddresses: string
  bodyPlain: string
  bodyHtmlSafe: string
  isHtml: boolean
  hasRemoteContent: boolean
  // remoteAllowed is true when remote content was rendered because the sender or
  // domain is trusted (or the global setting is on), so no banner is shown.
  remoteAllowed: boolean
  // remoteHosts lists the hosts blocked remote content would load from.
  remoteHosts: string[]
  attachments: Attachment[]
}

export interface MessageList {
  messages: MessageSummary[]
  total: number
}

export interface Attachment {
  id: number
  filename: string
  contentType: string
  sizeBytes: number
  inline: boolean
}

export interface Address {
  name: string
  email: string
}

export interface ComposeAttachment {
  filename: string
  contentType: string
  contentBase64: string
  inline: boolean
  contentId: string
}

export interface ComposeRequest {
  accountId: number
  to: Address[]
  cc: Address[]
  bcc: Address[]
  subject: string
  text: string
  html: string
  inReplyTo: string
  references: string[]
  attachments: ComposeAttachment[]
}

export interface Draft {
  id: number
  savedAt: string
  request: ComposeRequest
}

export interface OutboxRow {
  id: number
  accountId: number
  recipients: string[]
  state: string
  attempts: number
  lastError: string
  nextAttemptAt: string
  createdAt: string
}

export interface UIPrefs {
  theme: string
  accent: string
  density: string
  showMailboxBadge: boolean
  showDateTime: boolean
  showPgp: boolean
  showAuth: boolean
  toastPosition: string
  paneLocked: boolean
  sidebarWidth: number
  listWidth: number
  // sendDelaySeconds holds outgoing mail for this many seconds so the user can
  // undo. 0 disables the delay.
  sendDelaySeconds: number
  // flagHighlight controls how flagged rows stand out: flag, left, both, off.
  flagHighlight: string
  // showShortcutHints toggles inline keyboard shortcut chips (off by default).
  showShortcutHints: boolean
  // showAccountEmail shows the account email instead of its name in the sidebar.
  showAccountEmail: boolean
  // alwaysLoadImages disables remote-image blocking globally (off by default).
  alwaysLoadImages: boolean
  // avatarSource selects the sender-photo fallback chain: bimi_gravatar,
  // gravatar_bimi, or pfp (generated only). avatarStyle picks the generated
  // placeholder look: initials, mono, pixel, or geometric.
  avatarSource: string
  avatarStyle: string
  // multiSelectEnabled allows selecting several rows at once for bulk actions.
  multiSelectEnabled: boolean
  // showSelectedCount toggles the "N selected" count text in the selection bar.
  showSelectedCount: boolean
  // sidebarIndentGuides draws vertical guide lines for nested folders.
  sidebarIndentGuides: boolean
  // rowTemplate selects the list row layout: relaxed, comfortable, compact, single.
  rowTemplate: string
  // rowShowAvatar / rowShowSnippet are per-field overrides on the row template.
  rowShowAvatar: boolean
  rowShowSnippet: boolean
  // previewLines clamps the snippet to this many lines (where the template allows).
  previewLines: number
  // uiScale zooms the whole interface (string multiplier, '1' = 100%).
  uiScale: string
  // messageFontSize is the base font size (px) for rendered email content.
  messageFontSize: number
  // showFlaggedCount shows the count and bold styling on the sidebar Flagged view.
  showFlaggedCount: boolean
  // flagColorSync pushes color labels to the server as imap keywords.
  flagColorSync: boolean
  // showOfflineIndicator shows the little downloaded badge on pinned messages.
  showOfflineIndicator: boolean
  // swipe gestures on message rows (trackpad only).
  swipeEnabled: boolean
  swipeLeftAction: string
  swipeRightAction: string
  // composeVimMode enables vim keybindings in the compose editor.
  composeVimMode: boolean
  // downloadIncludeAttachments is the remembered default for the range download.
  downloadIncludeAttachments: boolean
  // appVimMode enables global vim-style navigation (h/j/k/l) for moving around
  // the app window itself, outside of compose.
  appVimMode: boolean
  // language is the ui locale code (en, de, fr, nl, es).
  language: string
  // lowPowerMode pauses periodic auto-sync, bulk downloads and address-book
  // rescans. autoSyncIntervalSeconds is how often a full sync pass runs on top
  // of the always-on imap idle push (0 disables it).
  lowPowerMode: boolean
  autoSyncIntervalSeconds: number
  // defaultEditorMode is the editor a new compose session starts in
  // (plaintext, markdown, or wysiwyg).
  defaultEditorMode: string
  // composeAutocomplete offers address-book suggestions while typing a
  // recipient. composeChips renders recipients as removable chips; when off,
  // the recipient fields fall back to a plain comma-separated text input.
  composeAutocomplete: boolean
  composeChips: boolean
  // updateCheckFrequency controls the automatic GitHub-releases update check:
  // 'off' (default), 'startup', 'weekly', or 'monthly'.
  updateCheckFrequency: string
}

// a harvested contact for compose autocomplete and the settings manager.
export interface AddressBookEntry {
  email: string
  name: string
  useCount: number
  lastUsed: string
  createdAt: string
}

// an attachment's bytes for the in-app previewer. data is base64; tooLarge is
// set (with empty data) when the file exceeds the preview cap.
export interface AttachmentContent {
  filename: string
  contentType: string
  sizeBytes: number
  data: string
  tooLarge: boolean
}

// the eight flag colors. index 0 means "no color"; 1..8 map to the palette in
// theme/flagcolors.ts and to imap $Label1..$Label8 when syncing is on.
export type FlagColor = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8

// swipe gesture actions for message rows.
export type SwipeAction = 'none' | 'delete' | 'read' | 'unread' | 'flag' | 'archive' | 'snooze'

// list row layouts, from most spacious to a single dense line.
export type RowTemplate = 'relaxed' | 'comfortable' | 'compact' | 'single'

// the six corners/edges a toast stack can anchor to.
export type ToastPosition =
  | 'top-left'
  | 'top-center'
  | 'top-right'
  | 'bottom-left'
  | 'bottom-center'
  | 'bottom-right'

export interface SettingResult {
  value: string
  found: boolean
}

// a reusable header/footer block. kind places it (top/bottom of a new message);
// format picks how the content is inserted into the compose body.
export interface Signature {
  id: number
  name: string
  kind: 'header' | 'footer'
  format: 'markdown' | 'html'
  content: string
}

// an account's default header/footer assignment. 0 means no default.
export interface AccountSignatures {
  headerId: number
  footerId: number
}

// autodiscovery result for the add-mailbox wizard.
export interface Discovered {
  imapHost: string
  imapPort: number
  smtpHost: string
  smtpPort: number
  oauth: boolean
  source: string
}

// the metadata the wizard collects to create an account. password is set for
// password auth; provider + clientId are set for oauth (per-user PKCE).
export interface AddAccountRequest {
  email: string
  displayName: string
  imapHost: string
  imapPort: number
  smtpHost: string
  smtpPort: number
  password: string
  provider: string
  clientId: string
  // optional oauth client secret for confidential-client app registrations
  // (some Microsoft Entra setups). empty keeps the default PKCE public flow.
  clientSecret: string
}

export interface TestConnectionRequest {
  email: string
  imapHost: string
  imapPort: number
  password: string
}

// folder roles mirror the backend's folderRole classification.
export type FolderRole =
  | 'inbox'
  | 'sent'
  | 'drafts'
  | 'trash'
  | 'junk'
  | 'archive'
  | 'normal'

// unified view keys mirror the backend view constants.
export type ViewKey = 'inbox' | 'flagged' | 'sent' | 'drafts' | 'archive' | 'junk' | 'trash'

// pgp status mirrors mailview.PGPStatus.
export type PGPStatus = 'none' | 'signed' | 'encrypted'

// auth status placeholder. only "unavailable" exists until the backend parses
// Authentication-Results headers (documented follow-up).
export type AuthStatus = 'unavailable'

// editor modes for the compose pane.
export type EditorMode = 'plaintext' | 'markdown' | 'wysiwyg'

// theme and density preference values.
export type ThemePref = 'system' | 'light' | 'dark'
export type DensityPref = 'compact' | 'medium' | 'luxe'

// Selection identifies what the message list is currently showing: either a
// unified cross-account view or a single account folder.
export type Selection =
  | { kind: 'view'; view: ViewKey; label: string }
  | { kind: 'folder'; folderId: number; accountId: number; label: string }
