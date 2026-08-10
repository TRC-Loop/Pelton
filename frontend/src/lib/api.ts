// api.ts is the single typed boundary between the svelte ui and the go backend.
// components call these functions, never window.go.* or the generated bindings
// directly, so the call sites stay typed and the generated layer can change
// shape without touching components.

import * as App from '../../wailsjs/go/desktop/App'
import { desktop } from '../../wailsjs/go/models'
import {
  isDemoActive,
  demoAccounts,
  demoFolders,
  demoViews,
  demoList,
  demoMessage,
  demoOutbox,
} from './demo'
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
  AddressBookEntry,
  AttachmentContent,
  ThemeInfo,
  ThemeApply,
  ThemeImportPreview,
  SaveThemeRequest,
  ThemeDraft,
  UserLocale,
  UserLocaleApply,
  ProxyConfig,
  MCPConfig,
  VirusTotalConfig,
  Verdict,
  MessageScan,
  View,
  Selection,
  FetchOlderResult,
  PGPKey,
} from './types'

// isDemoMode reports whether the app launched in the cosmetic --potatoes-are-nice
// screenshot mode; the frontend reads it once to switch the data layer to samples.
export function isDemoMode(): Promise<boolean> {
  return App.IsDemoMode()
}

// isDevMode reports whether the app is running against the separate dev data
// directory (PELTON_DEV), so the ui can show a persistent reminder that this
// isn't pointed at a real install.
export function isDevMode(): Promise<boolean> {
  return App.IsDevMode()
}

// isNightly reports whether this is an automated nightly build cut from the dev
// branch. Nightlies are untested and keep their own separate data directory, so
// the ui warns about it at every launch and marks the status bar.
export function isNightly(): Promise<boolean> {
  return App.IsNightly()
}

// consumePendingMailto returns the mailto: draft the app was launched with (if
// any) and clears it, so a reload does not reopen the same compose.
export function consumePendingMailto(): Promise<desktop.PendingMailtoDTO> {
  return App.ConsumePendingMailto()
}

// defaultMailClientStatus reports whether Pelton is the default mailto handler.
// known is false where the platform cannot answer reliably; the ui then shows
// nothing rather than guessing.
export function defaultMailClientStatus(): Promise<desktop.DefaultMailStatusDTO> {
  return App.DefaultMailClientStatus()
}

// setDefaultMailClient asks the OS to make Pelton the default mailto handler
// (a system sheet on macOS, the xdg association on Linux, the Settings page on
// Windows).
export function setDefaultMailClient(): Promise<void> {
  return App.SetDefaultMailClient()
}

// listAccounts returns every configured account.
export function listAccounts(): Promise<Account[]> {
  if (isDemoActive()) {
    return Promise.resolve(demoAccounts())
  }
  return App.ListAccounts()
}

// updateAccount persists edits to an account's display name and server settings.
export function updateAccount(req: {
  id: number
  displayName: string
  username: string
  imapHost: string
  imapPort: number
  smtpHost: string
  smtpPort: number
}): Promise<Account> {
  return App.UpdateAccount(new desktop.UpdateAccountRequest(req))
}

// deleteAccount removes an account, its cached mail and its keyring secret.
export function deleteAccount(id: number): Promise<void> {
  return App.DeleteAccount(id)
}

// listFolders returns one account's full mailbox tree with counts.
export function listFolders(accountId: number): Promise<Folder[]> {
  if (isDemoActive()) {
    return Promise.resolve(demoFolders(accountId))
  }
  return App.ListFolders(accountId)
}

// createFolder creates a mailbox on the server and returns the new folder.
// parentId 0 creates it at the root of the account. name is one level: nesting
// comes from parentId, not from embedding the server's delimiter in the name.
export function createFolder(accountId: number, parentId: number, name: string): Promise<Folder> {
  return App.CreateFolder(new desktop.CreateFolderRequest({ accountId, parentId, name }))
}

// renameFolder renames a mailbox in place, keeping its cached mail. Special
// mailboxes (INBOX, Sent, Drafts, Trash, Junk, Archive) are refused by the
// backend.
export function renameFolder(id: number, name: string): Promise<void> {
  return App.RenameFolder(id, name)
}

// deleteFolder deletes a mailbox, its subfolders and their mail, on the server
// as well as locally. Destructive and not undoable: confirm before calling.
export function deleteFolder(id: number): Promise<void> {
  return App.DeleteFolder(id)
}

// emptyTrash permanently deletes every message in a trash folder and resolves
// with how many it removed. The backend refuses any folder that is not the
// account's trash. Destructive and not undoable: confirm before calling.
export function emptyTrash(folderId: number): Promise<number> {
  return App.EmptyTrash(folderId)
}

// listUnifiedViews returns the cross-account views with aggregate counts.
export function listUnifiedViews(): Promise<UnifiedView[]> {
  if (isDemoActive()) {
    return Promise.resolve(demoViews())
  }
  return App.ListUnifiedViews()
}

// listFolderMessages reads a page of a single folder.
export function listFolderMessages(
  folderId: number,
  limit: number,
  offset: number,
): Promise<MessageList> {
  if (isDemoActive()) {
    return Promise.resolve(demoList())
  }
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
  if (isDemoActive()) {
    return Promise.resolve(demoList())
  }
  return App.ListMessages(
    new desktop.ListMessagesRequest({ kind: 'view', folderId: 0, view, limit, offset }),
  )
}

// listSavedViewMessages reads a page of a saved View (preset search).
export function listSavedViewMessages(
  viewId: number,
  limit: number,
  offset: number,
): Promise<MessageList> {
  if (isDemoActive()) {
    return Promise.resolve(demoList())
  }
  return App.ListMessages(
    new desktop.ListMessagesRequest({ kind: 'savedView', folderId: 0, view: '', viewId, limit, offset }),
  )
}

// fetchOlderMessages pulls the next batch of older mail from the server for a
// selection, for when the local cache runs out before the mailbox does. It is a
// network round trip, unlike the list* functions above which only read the
// cache. In demo mode there is no server, so it reports nothing fetched.
export function fetchOlderMessages(sel: Selection): Promise<FetchOlderResult> {
  if (isDemoActive()) {
    return Promise.resolve({ fetched: 0, hasOlder: false })
  }
  return App.FetchOlderMessages(
    new desktop.ListMessagesRequest({
      kind: sel.kind,
      folderId: sel.kind === 'folder' ? sel.folderId : 0,
      view: sel.kind === 'view' ? sel.view : '',
      viewId: sel.kind === 'savedView' ? sel.viewId : 0,
      limit: 0,
      offset: 0,
    }),
  )
}

// listViews returns every saved View with its eager-run counts.
export function listViews(): Promise<View[]> {
  if (isDemoActive()) {
    return Promise.resolve([])
  }
  return App.ListViews() as unknown as Promise<View[]>
}

// saveView creates (id 0) or updates a saved View and returns it with fresh counts.
export function saveView(view: View): Promise<View> {
  return App.SaveView(new desktop.ViewDTO(view)) as unknown as Promise<View>
}

// deleteView removes a saved View.
export function deleteView(id: number): Promise<void> {
  return App.DeleteView(id)
}

// reorderViews persists a new sidebar order for the given view ids.
export function reorderViews(orderedIds: number[]): Promise<void> {
  return App.ReorderViews(orderedIds)
}

// reorderFolders persists a new order for one group of sibling folders. Every
// id must share an account and a parent; the backend rejects a mixed group.
export function reorderFolders(orderedIds: number[]): Promise<void> {
  return App.ReorderFolders(orderedIds)
}

// reorderAccounts persists a new order for the sidebar's account sections.
export function reorderAccounts(orderedIds: number[]): Promise<void> {
  return App.ReorderAccounts(orderedIds)
}

// reorderUnifiedViews persists a new order for the unified views block, by key.
export function reorderUnifiedViews(keys: string[]): Promise<void> {
  return App.ReorderUnifiedViews(keys)
}

// listPinnedFolders returns the pinned folders across every account in the
// order of the sidebar's Pinned group.
export function listPinnedFolders(): Promise<Folder[]> {
  if (isDemoActive()) {
    return Promise.resolve([])
  }
  return App.ListPinnedFolders()
}

// setFolderPinned mirrors a folder into the Pinned group, or removes it from
// there. The folder stays in its own account's tree either way.
export function setFolderPinned(folderId: number, pinned: boolean): Promise<void> {
  return App.SetFolderPinned(folderId, pinned)
}

// reorderPinnedFolders persists a new order for the Pinned group.
export function reorderPinnedFolders(orderedIds: number[]): Promise<void> {
  return App.ReorderPinnedFolders(orderedIds)
}

// getMessage returns the full message with sanitized body and attachments.
export function getMessage(id: number): Promise<MessageDetail> {
  if (isDemoActive()) {
    return Promise.resolve(demoMessage(id))
  }
  return App.GetMessage(id) as unknown as Promise<MessageDetail>
}

// unsubscribeMessage carries out a message's advertised unsubscribe (the
// one-click POST or the mailto send); plain links are opened by the caller.
export function unsubscribeMessage(id: number): Promise<void> {
  return App.Unsubscribe(id)
}

// getMessageSource fetches a message's raw RFC 822 source on demand over imap.
export function getMessageSource(id: number): Promise<string> {
  return App.GetMessageSource(id)
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

// undoDelete reverses a pending delete while the message is still cached.
export function undoDelete(id: number): Promise<void> {
  return App.UndoDelete(id)
}

// ArchiveUndo is what undo-archive needs: the message's stable rfc Message-ID and
// the folder it came from. messageId is empty when the message had no Message-ID
// (undo not possible then).
export interface ArchiveUndo {
  messageId: string
  originalFolderId: number
}

// archiveMessage moves a message to its account's Archive folder on the server,
// returning the info needed to undo it.
export function archiveMessage(id: number): Promise<ArchiveUndo> {
  return App.ArchiveMessage(id)
}

// unarchiveMessage moves an archived message back to its original folder,
// locating it by rfc Message-ID.
export function unarchiveMessage(messageId: string, originalFolderId: number): Promise<void> {
  return App.UnarchiveMessage(messageId, originalFolderId)
}

// moveMessage moves a message to any folder of its account, returning undo info.
export function moveMessage(id: number, destFolderId: number): Promise<ArchiveUndo> {
  return App.MoveMessage(id, destFolderId)
}

// SearchRequest is a ranked search over cached mail: free text plus an optional
// date window. afterUnix/beforeUnix are unix seconds; 0 leaves that side open.
export interface SearchRequest {
  query: string
  afterUnix: number
  beforeUnix: number
  limit: number
  // field-scoped constraints from typed search chips (from:/to:/subject:).
  from: string
  to: string
  subject: string
  hasAttachment: boolean
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
  if (isDemoActive()) {
    return Promise.resolve(demoOutbox())
  }
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

// UpdateCheckResult mirrors the go DTO of the same name.
export interface UpdateCheckResult {
  checked: boolean
  available: boolean
  currentVersion: string
  latestVersion: string
  releaseUrl: string
  error: string
}

// checkForUpdates does an immediate GitHub-releases check (the "Check now"
// button in settings), regardless of the update_check_frequency setting.
export function checkForUpdates(): Promise<UpdateCheckResult> {
  return App.CheckForUpdates()
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

// allowRemoteForMessage permanently allows remote content for this one message,
// without trusting the whole sender or domain.
export function allowRemoteForMessage(messageId: number): Promise<void> {
  return App.AllowRemoteForMessage(messageId)
}

// ImageAllowEntry is one trusted sender or domain in the remote-image
// allowlist, with an example cached message when one exists.
export type ImageAllowEntry = desktop.ImageAllowEntryDTO

// listImageAllowlist returns every trusted sender and domain the user has
// allowed remote content for.
export function listImageAllowlist(): Promise<ImageAllowEntry[]> {
  return App.ListImageAllowlist()
}

// removeImageAllow removes a trusted sender or domain from the allowlist.
export function removeImageAllow(kind: 'sender' | 'domain', value: string): Promise<void> {
  return App.RemoveImageAllow(kind, value)
}

// listVIPSenders returns the addresses the user has marked as VIP (#126).
export function listVIPSenders(): Promise<string[]> {
  return App.ListVIPSenders()
}

// addVIPSender marks an address as VIP. The backend normalizes it to the bare
// lowercased address.
export function addVIPSender(address: string): Promise<void> {
  return App.AddVIPSender(address)
}

// removeVIPSender drops an address from the VIP list.
export function removeVIPSender(address: string): Promise<void> {
  return App.RemoveVIPSender(address)
}

// markSenderVIP adds a message's sender to the VIP list.
export function markSenderVIP(messageId: number): Promise<void> {
  return App.MarkSenderVIP(messageId)
}

// unmarkSenderVIP removes a message's sender from the VIP list.
export function unmarkSenderVIP(messageId: number): Promise<void> {
  return App.UnmarkSenderVIP(messageId)
}

// senderPhotos resolves the ordered list of remote photo candidates for a sender
// under the configured fallback chain. empty means "no network source"; the ui
// then draws a generated placeholder.
export function senderPhotos(email: string): Promise<string[]> {
  if (isDemoActive()) {
    // demo mode stays offline: fall back to the generated avatars, no network.
    return Promise.resolve([])
  }
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

// --- flag color, snooze, offline ---

// setFlagColor sets a message's color label (0 clears, 1..8 pick a color).
export function setFlagColor(id: number, color: number): Promise<void> {
  return App.SetFlagColor(id, color)
}

// downloadMessageOffline / removeOffline pin or unpin a single message.
export function downloadMessageOffline(id: number): Promise<void> {
  return App.DownloadMessageOffline(id)
}

export function removeOffline(id: number): Promise<void> {
  return App.RemoveOffline(id)
}

// snoozeMessage schedules a message to resurface at untilRFC3339; hideNow also
// hides it from the list until then.
export function snoozeMessage(id: number, untilRFC3339: string, hideNow: boolean): Promise<void> {
  return App.SnoozeMessage(id, untilRFC3339, hideNow)
}

export function unsnoozeMessage(id: number): Promise<void> {
  return App.UnsnoozeMessage(id)
}

// --- attachments (preview, save-all) ---

// readAttachment returns an attachment's bytes for the in-app previewer.
export function readAttachment(messageId: number, attachmentId: number): Promise<AttachmentContent> {
  return App.ReadAttachment(messageId, attachmentId)
}

// saveAllAttachments prompts for a directory and writes every attachment there,
// returning the directory (empty if cancelled).
export function saveAllAttachments(messageId: number): Promise<string> {
  return App.SaveAllAttachments(messageId)
}

// --- offline range download ---

// downloadRange downloads all mail since the start date that is not cached yet.
export function downloadRange(startDate: string, includeAttachments: boolean): Promise<void> {
  return App.DownloadRange(startDate, includeAttachments)
}

// cancelDownload stops a running bulk offline download and clears its resume
// marker so it does not restart on the next launch.
export function cancelDownload(): Promise<void> {
  return App.CancelDownload()
}

// --- import / export ---

// BackupInfo describes a picked backup file before importing it.
export type BackupInfo = desktop.BackupInfoDTO

// exportData writes the selected categories to a user-chosen json file,
// returning its path (empty if the save dialog was cancelled). credentialPassword,
// when non-empty, also encrypts and includes each exported mailbox's stored
// credential; pass an empty string to leave credentials out entirely.
export function exportData(categories: string[], credentialPassword: string): Promise<string> {
  return App.ExportData(categories, credentialPassword)
}

// inspectBackupFile opens a file picker and returns what the chosen backup
// holds. An empty path means the dialog was cancelled.
export function inspectBackupFile(): Promise<BackupInfo> {
  return App.InspectBackupFile()
}

// importData applies the selected categories from the backup file at path.
// credentialPassword, when non-empty, also decrypts and restores any mailbox
// credentials the file carries; a wrong password rejects with an error.
export function importData(path: string, categories: string[], credentialPassword: string): Promise<void> {
  return App.ImportData(path, categories, credentialPassword)
}

// systemColorScheme returns the OS dark/light preference ("dark" | "light"), or
// "" when it cannot be determined. Only meaningful on Linux, where WebKitGTK
// does not expose it to CSS prefers-color-scheme; elsewhere it returns "".
export function systemColorScheme(): Promise<string> {
  return App.SystemColorScheme()
}

// --- address book ---

export function searchAddresses(query: string, limit: number): Promise<AddressBookEntry[]> {
  return App.SearchAddresses(query, limit)
}

export function listAddresses(): Promise<AddressBookEntry[]> {
  return App.ListAddresses()
}

export function deleteAddress(email: string): Promise<void> {
  return App.DeleteAddress(email)
}

// --- window ---

// setWindowTitle updates the native window title to reflect context.
export function setWindowTitle(title: string): void {
  void App.SetWindowTitle(title)
}

// setWindowTheme matches the native window chrome (the Windows caption bar) to
// the resolved ui theme. No-op on macOS/Linux.
export function setWindowTheme(dark: boolean): void {
  void App.SetWindowTheme(dark)
}

// setMailActionsEnabled greys out or restores the native Mail menu's message
// actions; the app calls it as the open message changes so those items are only
// selectable while a message is open.
export function setMailActionsEnabled(enabled: boolean): void {
  void App.SetMailActionsEnabled(enabled)
}

// getUIPrefs returns all ui preferences with defaults applied server-side. The
// backend types viewsPlacement as a plain string; it only ever emits the
// ViewsPlacement values, so the cast is safe.
export function getUIPrefs(): Promise<UIPrefs> {
  return App.GetUIPrefs() as unknown as Promise<UIPrefs>
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
  viewsPlacement: 'views_placement',
  flagColorSync: 'flag_color_sync',
  showOfflineIndicator: 'show_offline_indicator',
  swipeEnabled: 'swipe_enabled',
  swipeLeftAction: 'swipe_left_action',
  swipeRightAction: 'swipe_right_action',
  composeVimMode: 'compose_vim_mode',
  downloadIncludeAttachments: 'download_include_attachments',
  appVimMode: 'app_vim_mode',
  language: 'language',
  lowPowerMode: 'low_power_mode',
  autoSyncIntervalSeconds: 'auto_sync_interval_seconds',
  defaultEditorMode: 'default_editor_mode',
  composeAutocomplete: 'compose_autocomplete',
  composeChips: 'compose_chips',
  updateCheckFrequency: 'update_check_frequency',
  emptyStateImage: 'empty_state_image',
  emptyStateFullscreen: 'empty_state_fullscreen',
  cornerStyle: 'corner_style',
  themeId: 'theme_id',
  menuBarInApp: 'menu_bar_in_app',
  menuBarNativeMinimal: 'menu_bar_native_minimal',
  menuBarIcons: 'menu_bar_icons',
  menuBarLayout: 'menu_bar_layout',
  menuBarNewItems: 'menu_bar_new_items',
  timeFormat: 'time_format',
  reduceMotion: 'reduce_motion',
  themeDarkStart: 'theme_dark_start',
  themeDarkEnd: 'theme_dark_end',
  bodyFont: 'body_font',
  uiFont: 'ui_font',
  monoFont: 'mono_font',
  notifyNewMail: 'notify_new_mail',
  verboseSync: 'verbose_sync',
  closeAction: 'close_button_action',
  syncMessageLimit: 'sync_message_limit',
  syncAutoBackfill: 'sync_auto_backfill',
  startupSelection: 'startup_selection',
  lastSelection: 'last_selection',
  liabilityAccepted: 'liability_accepted',
} as const

// listSystemFonts returns the installed font family names for the body font
// dropdown (cached backend-side after the first scan).
export function listSystemFonts(): Promise<string[]> {
  return App.ListSystemFonts().then((f) => f ?? [])
}

// --- pgp keys (#192) ---

// listPGPKeys returns every imported key, private keys first.
export function listPGPKeys(): Promise<PGPKey[]> {
  return App.ListPGPKeys() as unknown as Promise<PGPKey[]>
}

// importPGPKey opens a file picker and imports the keys the file holds.
// Resolves with an empty list when the dialog was cancelled.
export function importPGPKey(): Promise<PGPKey[]> {
  return App.ImportPGPKey().then((keys) => (keys ?? []) as unknown as PGPKey[])
}

// deletePGPKey removes a key from both rings and forgets its passphrase.
export function deletePGPKey(fingerprint: string): Promise<void> {
  return App.DeletePGPKey(fingerprint)
}

// exportPGPKey writes a key to a file the user picks, returning the path or ''
// if cancelled. includePrivate writes the private half, which is the only
// backup path for it: keys are deliberately excluded from the backup archive.
export function exportPGPKey(fingerprint: string, includePrivate: boolean): Promise<string> {
  return App.ExportPGPKey(fingerprint, includePrivate)
}

// unlockPGPKey verifies a passphrase and holds it for the session. remember
// also stores it in the os keyring so it survives a restart.
export function unlockPGPKey(
  fingerprint: string,
  passphrase: string,
  remember: boolean,
): Promise<void> {
  return App.UnlockPGPKey(fingerprint, passphrase, remember)
}

// forgetPGPPassphrase drops a passphrase from the session and the keyring.
export function forgetPGPPassphrase(fingerprint: string): Promise<void> {
  return App.ForgetPGPPassphrase(fingerprint)
}

// getAccountPGPKey returns the fingerprint an account signs with, or ''.
export function getAccountPGPKey(accountId: number): Promise<string> {
  return App.GetAccountPGPKey(accountId)
}

// setAccountPGPKey pins the signing key for an account; '' clears it and falls
// back to matching the account address against the keys' user ids.
export function setAccountPGPKey(accountId: number, fingerprint: string): Promise<void> {
  return App.SetAccountPGPKey(accountId, fingerprint)
}

// --- custom themes ---

// listThemes returns every installed custom theme for the settings gallery.
export function listThemes(): Promise<ThemeInfo[]> {
  return App.ListThemes() as Promise<ThemeInfo[]>
}

// getThemeApply loads an installed theme in apply form (base, validated token
// overrides, concatenated css with bundled assets inlined, icon svgs).
export function getThemeApply(id: string): Promise<ThemeApply> {
  return App.GetThemeApply(id) as Promise<ThemeApply>
}

// previewThemeImport opens a file picker for a .peltontheme container and
// returns the read-before-import view. canceled is true when dismissed;
// nothing is installed yet.
export function previewThemeImport(): Promise<ThemeImportPreview> {
  return App.PreviewThemeImport() as Promise<ThemeImportPreview>
}

// confirmThemeImport installs a previewed container. allowRemote keeps the
// css's network references; false strips them before anything hits disk.
// importTokens and importCSS are the parts choice: a deselected part is
// dropped from the container before it is written.
export function confirmThemeImport(
  path: string,
  allowRemote: boolean,
  importTokens: boolean,
  importCSS: boolean,
): Promise<ThemeInfo> {
  return App.ConfirmThemeImport(path, allowRemote, importTokens, importCSS) as Promise<ThemeInfo>
}

// deleteTheme removes an installed theme (and resets the selection if it was
// active).
export function deleteTheme(id: string): Promise<void> {
  return App.DeleteTheme(id)
}

// exportTheme zips an installed theme back into a shareable .peltontheme via
// a save dialog, returning the chosen path ('' if cancelled).
export function exportTheme(id: string): Promise<string> {
  return App.ExportTheme(id)
}

// saveCustomTheme validates and writes a palette-editor theme as a
// .peltontheme file in the themes folder, returning its gallery info.
export function saveCustomTheme(req: SaveThemeRequest): Promise<ThemeInfo> {
  return App.SaveCustomTheme(new desktop.SaveThemeRequest(req)) as Promise<ThemeInfo>
}

// getThemeDraft loads an installed theme back into editor form: metadata,
// tokens and the raw source of the editor's own stylesheet.
export function getThemeDraft(id: string): Promise<ThemeDraft> {
  return App.GetThemeDraft(id) as Promise<ThemeDraft>
}

// openThemesFolder shows the themes folder in the system file manager.
export function openThemesFolder(): Promise<void> {
  return App.OpenThemesFolder()
}

// --- custom languages ---

// listUserLocales returns every valid language file in the locales folder.
export function listUserLocales(): Promise<UserLocale[]> {
  return App.ListUserLocales() as Promise<UserLocale[]>
}

// getUserLocale loads one custom language in apply form (base + strings).
export function getUserLocale(id: string): Promise<UserLocaleApply> {
  return App.GetUserLocale(id) as Promise<UserLocaleApply>
}

// openLocalesFolder shows the locales folder in the system file manager.
export function openLocalesFolder(): Promise<void> {
  return App.OpenLocalesFolder()
}

// saveLocaleTemplate writes a translation template file via a save dialog,
// returning the chosen path ('' if cancelled).
export function saveLocaleTemplate(content: string): Promise<string> {
  return App.SaveLocaleTemplate(content)
}

// getProxyConfig returns the current outbound proxy preference (without the
// stored password; hasPassword reports whether one is set).
export function getProxyConfig(): Promise<ProxyConfig> {
  return App.GetProxyConfig() as Promise<ProxyConfig>
}

// setProxyConfig validates, persists and applies a proxy preference. Leave
// password empty with hasPassword true to keep the stored one.
export function setProxyConfig(cfg: ProxyConfig): Promise<void> {
  return App.SetProxyConfig(new desktop.ProxyConfigDTO(cfg))
}

// testProxy dials a well-known endpoint through the given settings so the ui can
// confirm the proxy works before saving. Resolves on success.
export function testProxy(cfg: ProxyConfig): Promise<void> {
  return App.TestProxy(new desktop.ProxyConfigDTO(cfg))
}

// getMCPConfig returns the read-only MCP server's state: enabled, loopback url,
// bearer token and whether it is listening.
export function getMCPConfig(): Promise<MCPConfig> {
  return App.GetMCPConfig() as Promise<MCPConfig>
}

// setMCPEnabled turns the MCP server on or off, generating a token on first
// enable so the endpoint is never unauthenticated.
export function setMCPEnabled(enabled: boolean): Promise<void> {
  return App.SetMCPEnabled(enabled)
}

// setMCPPort changes the loopback port (1024-65535) and restarts the server if
// it is running.
export function setMCPPort(port: number): Promise<void> {
  return App.SetMCPPort(port)
}

// regenerateMCPToken issues a fresh bearer token, invalidating the old one, and
// returns it for display.
export function regenerateMCPToken(): Promise<string> {
  return App.RegenerateMCPToken()
}

// getVirusTotalConfig returns the VirusTotal integration's settings: whether it
// is on, whether a key is stored, and the two auto-scan toggles.
export function getVirusTotalConfig(): Promise<VirusTotalConfig> {
  return App.GetVirusTotalConfig() as Promise<VirusTotalConfig>
}

// setVirusTotalEnabled turns the integration on or off. Turning it off also
// discards every cached verdict and resets both auto-scan toggles.
export function setVirusTotalEnabled(enabled: boolean): Promise<void> {
  return App.SetVirusTotalEnabled(enabled)
}

// setVirusTotalApiKey stores the api key in the os keyring, or clears it (and
// the cached verdicts it produced) when given an empty string.
export function setVirusTotalApiKey(apiKey: string): Promise<void> {
  return App.SetVirusTotalAPIKey(apiKey)
}

// setVirusTotalAutoScanLinks turns automatic link scanning on or off.
export function setVirusTotalAutoScanLinks(enabled: boolean): Promise<void> {
  return App.SetVirusTotalAutoScanLinks(enabled)
}

// setVirusTotalAutoScanAttachments turns automatic attachment scanning on or off.
export function setVirusTotalAutoScanAttachments(enabled: boolean): Promise<void> {
  return App.SetVirusTotalAutoScanAttachments(enabled)
}

// scanUrl looks one link up on VirusTotal, answering from the local cache when
// it can. This is the on-demand path behind the link context menu.
export function scanUrl(url: string): Promise<Verdict> {
  return App.ScanURL(url) as Promise<Verdict>
}

// scanAttachment looks one attachment up by the sha-256 of its bytes. The file
// itself is never uploaded, so an unrecognised file stays 'unknown'.
export function scanAttachment(messageId: number, attachmentId: number): Promise<Verdict> {
  return App.ScanAttachment(messageId, attachmentId) as Promise<Verdict>
}

// scanMessage scans a whole message's links, attachments, or both. Results come
// back per target, so one failed lookup does not lose the others.
export function scanMessage(messageId: number, links: boolean, attachments: boolean): Promise<MessageScan> {
  return App.ScanMessage(messageId, links, attachments) as Promise<MessageScan>
}
