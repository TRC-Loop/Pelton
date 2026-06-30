// api.ts is the single typed boundary between the svelte ui and the go backend.
// components call these functions, never window.go.* or the generated bindings
// directly, so the call sites stay typed and the generated layer can change
// shape without touching components.

import * as App from '../../wailsjs/go/desktop/App'
import { desktop } from '../../wailsjs/go/models'
import type {
  Account,
  Folder,
  UnifiedView,
  MessageList,
  MessageDetail,
  MessageSummary,
  ComposeRequest,
  Draft,
  OutboxRow,
  UIPrefs,
  ViewKey,
  Discovered,
  AddAccountRequest,
  TestConnectionRequest,
  Signature,
  AccountSignatures,
} from './types'

// listAccounts returns every configured account.
export function listAccounts(): Promise<Account[]> {
  return App.ListAccounts()
}

// listFolders returns one account's full mailbox tree with counts.
export function listFolders(accountId: number): Promise<Folder[]> {
  return App.ListFolders(accountId)
}

// listUnifiedViews returns the cross-account views with aggregate counts.
export function listUnifiedViews(): Promise<UnifiedView[]> {
  return App.ListUnifiedViews()
}

// listFolderMessages reads a page of a single folder.
export function listFolderMessages(
  folderId: number,
  limit: number,
  offset: number,
): Promise<MessageList> {
  return App.ListMessages(
    new desktop.ListMessagesRequest({ kind: 'folder', folderId, view: '', limit, offset }),
  )
}

// listViewMessages reads a page of a unified view.
export function listViewMessages(
  view: ViewKey,
  limit: number,
  offset: number,
): Promise<MessageList> {
  return App.ListMessages(
    new desktop.ListMessagesRequest({ kind: 'view', folderId: 0, view, limit, offset }),
  )
}

// getMessage returns the full message with sanitized body and attachments.
export function getMessage(id: number): Promise<MessageDetail> {
  return App.GetMessage(id)
}

// getMessageHtml re-renders a body with the chosen remote-image policy.
export function getMessageHtml(id: number, allowRemote: boolean): Promise<string> {
  return App.GetMessageHTML(id, allowRemote)
}

// setSeen / setFlagged toggle a flag and queue the change for sync.
export function setSeen(id: number, seen: boolean): Promise<void> {
  return App.SetSeen(id, seen)
}

export function setFlagged(id: number, flagged: boolean): Promise<void> {
  return App.SetFlagged(id, flagged)
}

// deleteMessage marks a message for server-side deletion on the next sync.
export function deleteMessage(id: number): Promise<void> {
  return App.DeleteMessage(id)
}

// SearchRequest is a ranked search over cached mail: free text plus an optional
// date window. afterUnix/beforeUnix are unix seconds; 0 leaves that side open.
export interface SearchRequest {
  query: string
  afterUnix: number
  beforeUnix: number
  limit: number
}

// search runs a ranked, typo-tolerant search and returns matching summaries in
// relevance order.
export function search(req: SearchRequest): Promise<MessageSummary[]> {
  return App.Search(new desktop.SearchRequestDTO(req))
}

// saveAttachment prompts for a path and writes the file, returning the path or
// an empty string if the user cancelled.
export function saveAttachment(messageId: number, attachmentId: number): Promise<string> {
  return App.SaveAttachment(messageId, attachmentId)
}

// sendMessage enqueues a message in the durable outbox. the plain request is
// wrapped back into the generated class the binding expects.
export function sendMessage(req: ComposeRequest): Promise<number> {
  return App.SendMessage(new desktop.ComposeRequest(req))
}

// saveDraft stores a compose request as a local draft, returning its id.
export function saveDraft(id: number, req: ComposeRequest): Promise<number> {
  return App.SaveDraft(id, new desktop.ComposeRequest(req))
}

export function listDrafts(): Promise<Draft[]> {
  return App.ListDrafts()
}

export function deleteDraft(id: number): Promise<void> {
  return App.DeleteDraft(id)
}

// listOutbox returns the current outbox contents.
export function listOutbox(): Promise<OutboxRow[]> {
  return App.ListOutbox()
}

// triggerSync runs one sync pass on demand.
export function triggerSync(): Promise<void> {
  return App.TriggerSync()
}

// appVersion returns the build version string (injected via ldflags), shown in
// the about section.
export function appVersion(): Promise<string> {
  return App.AppVersion()
}

// cancelSend pulls a still-queued message back out of the outbox during its
// undo-send delay window, resolving true when it was cancelled in time.
export function cancelSend(id: number): Promise<boolean> {
  return App.CancelSend(id)
}

// clearSentOutbox prunes rows already marked sent after the ui has shown the
// brief sent confirmation.
export function clearSentOutbox(): Promise<void> {
  return App.ClearSentOutbox()
}

// trustSenderImages permanently allows remote content from a message's sender.
export function trustSenderImages(messageId: number): Promise<void> {
  return App.TrustSenderImages(messageId)
}

// allowDomainImages permanently allows remote content from a sender's domain.
export function allowDomainImages(messageId: number): Promise<void> {
  return App.AllowDomainImages(messageId)
}

// senderPhotos resolves the ordered list of remote photo candidates for a sender
// under the configured fallback chain. empty means "no network source"; the ui
// then draws a generated placeholder.
export function senderPhotos(email: string): Promise<string[]> {
  return App.SenderPhotos(email)
}

// exportMessagePrintView opens a print-ready view of a message in the system
// browser, where it can be saved as a pdf or printed.
export function exportMessagePrintView(id: number): Promise<void> {
  return App.ExportMessagePrintView(id)
}

// LicenseEntry is one third-party dependency's license record.
export interface LicenseEntry {
  group: string
  name: string
  license: string
  text: string
}

// licenses returns the embedded third-party license manifest. it is fetched only
// when the about section's licenses view is opened, so the text stays out of the
// main bundle.
export async function licenses(): Promise<LicenseEntry[]> {
  const raw = await App.Licenses()
  try {
    return JSON.parse(raw) as LicenseEntry[]
  } catch {
    return []
  }
}

// programLicense returns Pelton's own license text (GPL-3.0).
export function programLicense(): Promise<string> {
  return App.ProgramLicense()
}

// --- add-mailbox wizard ---

// discoverConfig resolves likely imap/smtp settings for an email address.
export function discoverConfig(email: string): Promise<Discovered> {
  return App.DiscoverConfig(email)
}

// listOAuthProviders returns supported oauth provider keys mapped to labels.
export function listOAuthProviders(): Promise<Record<string, string>> {
  return App.ListOAuthProviders()
}

// testConnection verifies imap credentials by logging in. Resolves on success.
export function testConnection(req: TestConnectionRequest): Promise<void> {
  return App.TestConnection(new desktop.TestConnectionRequest(req))
}

// addPasswordAccount creates a password-authenticated account (stores the
// password in the keyring, discovers folders, syncs).
export function addPasswordAccount(req: AddAccountRequest): Promise<Account> {
  return App.AddPasswordAccount(new desktop.AddAccountRequest(req))
}

// addOAuthAccount runs the interactive PKCE flow then creates the account.
export function addOAuthAccount(req: AddAccountRequest): Promise<Account> {
  return App.AddOAuthAccount(new desktop.AddAccountRequest(req))
}

// --- signatures (header/footer blocks) ---

// listSignatures returns every signature block. the generated dto types kind as
// a plain string; we narrow back to the Signature union at this boundary.
export function listSignatures(): Promise<Signature[]> {
  return App.ListSignatures() as Promise<Signature[]>
}

// saveSignature creates the block when id is 0, otherwise updates it; resolves to
// the stored block (with its id).
export function saveSignature(s: Signature): Promise<Signature> {
  return App.SaveSignature(new desktop.SignatureDTO(s)) as Promise<Signature>
}

// deleteSignature removes a block; accounts defaulting to it have the slot cleared.
export function deleteSignature(id: number): Promise<void> {
  return App.DeleteSignature(id)
}

// getAccountSignatures returns an account's default header/footer ids (0 = none).
export function getAccountSignatures(accountId: number): Promise<AccountSignatures> {
  return App.GetAccountSignatures(accountId)
}

// setAccountSignatures sets an account's default header/footer (0 clears a slot).
export function setAccountSignatures(
  accountId: number,
  headerId: number,
  footerId: number,
): Promise<void> {
  return App.SetAccountSignatures(accountId, headerId, footerId)
}

// getUIPrefs returns all ui preferences with defaults applied server-side.
export function getUIPrefs(): Promise<UIPrefs> {
  return App.GetUIPrefs()
}

// setSetting writes a single preference by key.
export function setSetting(key: string, value: string): Promise<void> {
  return App.SetSetting(key, value)
}

// getSetting reads a single raw setting. found is false when the key was never
// written, so callers can apply their own default.
export function getSetting(key: string): Promise<{ value: string; found: boolean }> {
  return App.GetSetting(key)
}

// setting keys shared with the backend (bind_settings.go). centralized so the
// stores never sprinkle raw strings.
export const SettingKeys = {
  theme: 'theme',
  accent: 'accent',
  density: 'density',
  showMailboxBadge: 'show_mailbox_badge',
  showDateTime: 'show_datetime',
  showPgp: 'show_pgp',
  showAuth: 'show_auth',
  editorMode: 'editor_mode',
  toastPosition: 'toast_position',
  paneLocked: 'pane_locked',
  sidebarWidth: 'sidebar_width',
  listWidth: 'list_width',
  sendDelay: 'send_delay_seconds',
  flagHighlight: 'flag_highlight',
  shortcutHints: 'show_shortcut_hints',
  accountEmail: 'show_account_email',
  onboarded: 'onboarding_complete',
  alwaysLoadImages: 'remote_images_always',
  avatarSource: 'avatar_source',
  avatarStyle: 'avatar_style',
  multiSelectEnabled: 'multi_select_enabled',
  showSelectedCount: 'show_selected_count',
  sidebarIndentGuides: 'sidebar_indent_guides',
  rowTemplate: 'row_template',
  rowShowAvatar: 'row_show_avatar',
  rowShowSnippet: 'row_show_snippet',
  previewLines: 'preview_lines',
  uiScale: 'ui_scale',
  messageFontSize: 'message_font_size',
  showFlaggedCount: 'show_flagged_count',
} as const
