export namespace desktop {
	
	export class AccountDTO {
	    id: number;
	    email: string;
	    displayName: string;
	    imapHost: string;
	    smtpHost: string;
	
	    static createFrom(source: any = {}) {
	        return new AccountDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.email = source["email"];
	        this.displayName = source["displayName"];
	        this.imapHost = source["imapHost"];
	        this.smtpHost = source["smtpHost"];
	    }
	}
	export class AccountSignaturesDTO {
	    headerId: number;
	    footerId: number;
	
	    static createFrom(source: any = {}) {
	        return new AccountSignaturesDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.headerId = source["headerId"];
	        this.footerId = source["footerId"];
	    }
	}
	export class AddAccountRequest {
	    email: string;
	    displayName: string;
	    imapHost: string;
	    imapPort: number;
	    smtpHost: string;
	    smtpPort: number;
	    password: string;
	    provider: string;
	    clientId: string;
	    clientSecret: string;
	
	    static createFrom(source: any = {}) {
	        return new AddAccountRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.email = source["email"];
	        this.displayName = source["displayName"];
	        this.imapHost = source["imapHost"];
	        this.imapPort = source["imapPort"];
	        this.smtpHost = source["smtpHost"];
	        this.smtpPort = source["smtpPort"];
	        this.password = source["password"];
	        this.provider = source["provider"];
	        this.clientId = source["clientId"];
	        this.clientSecret = source["clientSecret"];
	    }
	}
	export class AddressBookEntryDTO {
	    email: string;
	    name: string;
	    useCount: number;
	    lastUsed: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new AddressBookEntryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.email = source["email"];
	        this.name = source["name"];
	        this.useCount = source["useCount"];
	        this.lastUsed = source["lastUsed"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class AddressDTO {
	    name: string;
	    email: string;
	
	    static createFrom(source: any = {}) {
	        return new AddressDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.email = source["email"];
	    }
	}
	export class ArchiveUndoDTO {
	    messageId: string;
	    originalFolderId: number;
	
	    static createFrom(source: any = {}) {
	        return new ArchiveUndoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.messageId = source["messageId"];
	        this.originalFolderId = source["originalFolderId"];
	    }
	}
	export class AttachmentContentDTO {
	    filename: string;
	    contentType: string;
	    sizeBytes: number;
	    data: string;
	    tooLarge: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AttachmentContentDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.contentType = source["contentType"];
	        this.sizeBytes = source["sizeBytes"];
	        this.data = source["data"];
	        this.tooLarge = source["tooLarge"];
	    }
	}
	export class AttachmentDTO {
	    id: number;
	    filename: string;
	    contentType: string;
	    sizeBytes: number;
	    inline: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AttachmentDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.filename = source["filename"];
	        this.contentType = source["contentType"];
	        this.sizeBytes = source["sizeBytes"];
	        this.inline = source["inline"];
	    }
	}
	export class ComposeAttachment {
	    filename: string;
	    contentType: string;
	    contentBase64: string;
	    inline: boolean;
	    contentId: string;
	
	    static createFrom(source: any = {}) {
	        return new ComposeAttachment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.contentType = source["contentType"];
	        this.contentBase64 = source["contentBase64"];
	        this.inline = source["inline"];
	        this.contentId = source["contentId"];
	    }
	}
	export class ComposeRequest {
	    accountId: number;
	    to: AddressDTO[];
	    cc: AddressDTO[];
	    bcc: AddressDTO[];
	    subject: string;
	    text: string;
	    html: string;
	    inReplyTo: string;
	    references: string[];
	    attachments: ComposeAttachment[];
	
	    static createFrom(source: any = {}) {
	        return new ComposeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.accountId = source["accountId"];
	        this.to = this.convertValues(source["to"], AddressDTO);
	        this.cc = this.convertValues(source["cc"], AddressDTO);
	        this.bcc = this.convertValues(source["bcc"], AddressDTO);
	        this.subject = source["subject"];
	        this.text = source["text"];
	        this.html = source["html"];
	        this.inReplyTo = source["inReplyTo"];
	        this.references = source["references"];
	        this.attachments = this.convertValues(source["attachments"], ComposeAttachment);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConfigSyncFolderPeekDTO {
	    hasExistingData: boolean;
	    accountEmails: string[];
	    modifiedUnix: number;
	
	    static createFrom(source: any = {}) {
	        return new ConfigSyncFolderPeekDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasExistingData = source["hasExistingData"];
	        this.accountEmails = source["accountEmails"];
	        this.modifiedUnix = source["modifiedUnix"];
	    }
	}
	export class ConfigSyncStatusDTO {
	    enabled: boolean;
	    mode: string;
	    path: string;
	    syncSettings: boolean;
	    emailScope: string;
	    lastSyncUnix: number;
	    lastError: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigSyncStatusDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.mode = source["mode"];
	        this.path = source["path"];
	        this.syncSettings = source["syncSettings"];
	        this.emailScope = source["emailScope"];
	        this.lastSyncUnix = source["lastSyncUnix"];
	        this.lastError = source["lastError"];
	    }
	}
	export class DiscoveredDTO {
	    imapHost: string;
	    imapPort: number;
	    smtpHost: string;
	    smtpPort: number;
	    oauth: boolean;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new DiscoveredDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imapHost = source["imapHost"];
	        this.imapPort = source["imapPort"];
	        this.smtpHost = source["smtpHost"];
	        this.smtpPort = source["smtpPort"];
	        this.oauth = source["oauth"];
	        this.source = source["source"];
	    }
	}
	export class DownloadEstimateDTO {
	    messageCount: number;
	    totalBytes: number;
	
	    static createFrom(source: any = {}) {
	        return new DownloadEstimateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.messageCount = source["messageCount"];
	        this.totalBytes = source["totalBytes"];
	    }
	}
	export class DraftDTO {
	    id: number;
	    savedAt: string;
	    request: ComposeRequest;
	
	    static createFrom(source: any = {}) {
	        return new DraftDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.savedAt = source["savedAt"];
	        this.request = this.convertValues(source["request"], ComposeRequest);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FolderDTO {
	    id: number;
	    accountId: number;
	    name: string;
	    imapPath: string;
	    delimiter: string;
	    parentId?: number;
	    role: string;
	    unreadCount: number;
	    totalCount: number;
	    attributes: string[];
	
	    static createFrom(source: any = {}) {
	        return new FolderDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.accountId = source["accountId"];
	        this.name = source["name"];
	        this.imapPath = source["imapPath"];
	        this.delimiter = source["delimiter"];
	        this.parentId = source["parentId"];
	        this.role = source["role"];
	        this.unreadCount = source["unreadCount"];
	        this.totalCount = source["totalCount"];
	        this.attributes = source["attributes"];
	    }
	}
	export class ListMessagesRequest {
	    kind: string;
	    folderId: number;
	    view: string;
	    limit: number;
	    offset: number;
	
	    static createFrom(source: any = {}) {
	        return new ListMessagesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.folderId = source["folderId"];
	        this.view = source["view"];
	        this.limit = source["limit"];
	        this.offset = source["offset"];
	    }
	}
	export class MessageDetailDTO {
	    id: number;
	    accountId: number;
	    folderId: number;
	    accountEmail: string;
	    folderName: string;
	    subject: string;
	    fromName: string;
	    fromAddress: string;
	    snippet: string;
	    date: string;
	    seen: boolean;
	    flagged: boolean;
	    hasAttachments: boolean;
	    pgp: string;
	    auth: string;
	    flagColor: number;
	    offline: boolean;
	    snoozeUntil: string;
	    toAddresses: string;
	    ccAddresses: string;
	    bodyPlain: string;
	    bodyHtmlSafe: string;
	    isHtml: boolean;
	    hasRemoteContent: boolean;
	    remoteAllowed: boolean;
	    remoteHosts: string[];
	    attachments: AttachmentDTO[];
	
	    static createFrom(source: any = {}) {
	        return new MessageDetailDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.accountId = source["accountId"];
	        this.folderId = source["folderId"];
	        this.accountEmail = source["accountEmail"];
	        this.folderName = source["folderName"];
	        this.subject = source["subject"];
	        this.fromName = source["fromName"];
	        this.fromAddress = source["fromAddress"];
	        this.snippet = source["snippet"];
	        this.date = source["date"];
	        this.seen = source["seen"];
	        this.flagged = source["flagged"];
	        this.hasAttachments = source["hasAttachments"];
	        this.pgp = source["pgp"];
	        this.auth = source["auth"];
	        this.flagColor = source["flagColor"];
	        this.offline = source["offline"];
	        this.snoozeUntil = source["snoozeUntil"];
	        this.toAddresses = source["toAddresses"];
	        this.ccAddresses = source["ccAddresses"];
	        this.bodyPlain = source["bodyPlain"];
	        this.bodyHtmlSafe = source["bodyHtmlSafe"];
	        this.isHtml = source["isHtml"];
	        this.hasRemoteContent = source["hasRemoteContent"];
	        this.remoteAllowed = source["remoteAllowed"];
	        this.remoteHosts = source["remoteHosts"];
	        this.attachments = this.convertValues(source["attachments"], AttachmentDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MessageSummaryDTO {
	    id: number;
	    accountId: number;
	    folderId: number;
	    accountEmail: string;
	    folderName: string;
	    subject: string;
	    fromName: string;
	    fromAddress: string;
	    snippet: string;
	    date: string;
	    seen: boolean;
	    flagged: boolean;
	    hasAttachments: boolean;
	    pgp: string;
	    auth: string;
	    flagColor: number;
	    offline: boolean;
	    snoozeUntil: string;
	
	    static createFrom(source: any = {}) {
	        return new MessageSummaryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.accountId = source["accountId"];
	        this.folderId = source["folderId"];
	        this.accountEmail = source["accountEmail"];
	        this.folderName = source["folderName"];
	        this.subject = source["subject"];
	        this.fromName = source["fromName"];
	        this.fromAddress = source["fromAddress"];
	        this.snippet = source["snippet"];
	        this.date = source["date"];
	        this.seen = source["seen"];
	        this.flagged = source["flagged"];
	        this.hasAttachments = source["hasAttachments"];
	        this.pgp = source["pgp"];
	        this.auth = source["auth"];
	        this.flagColor = source["flagColor"];
	        this.offline = source["offline"];
	        this.snoozeUntil = source["snoozeUntil"];
	    }
	}
	export class MessageListDTO {
	    messages: MessageSummaryDTO[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new MessageListDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.messages = this.convertValues(source["messages"], MessageSummaryDTO);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class OutboxRowDTO {
	    id: number;
	    accountId: number;
	    recipients: string[];
	    state: string;
	    attempts: number;
	    lastError: string;
	    nextAttemptAt: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new OutboxRowDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.accountId = source["accountId"];
	        this.recipients = source["recipients"];
	        this.state = source["state"];
	        this.attempts = source["attempts"];
	        this.lastError = source["lastError"];
	        this.nextAttemptAt = source["nextAttemptAt"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class SearchRequestDTO {
	    query: string;
	    afterUnix: number;
	    beforeUnix: number;
	    limit: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.afterUnix = source["afterUnix"];
	        this.beforeUnix = source["beforeUnix"];
	        this.limit = source["limit"];
	    }
	}
	export class SettingResult {
	    value: string;
	    found: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SettingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.found = source["found"];
	    }
	}
	export class SignatureDTO {
	    id: number;
	    name: string;
	    kind: string;
	    format: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new SignatureDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.format = source["format"];
	        this.content = source["content"];
	    }
	}
	export class TestConnectionRequest {
	    email: string;
	    imapHost: string;
	    imapPort: number;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new TestConnectionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.email = source["email"];
	        this.imapHost = source["imapHost"];
	        this.imapPort = source["imapPort"];
	        this.password = source["password"];
	    }
	}
	export class UIPrefsDTO {
	    theme: string;
	    accent: string;
	    density: string;
	    showMailboxBadge: boolean;
	    showDateTime: boolean;
	    showPgp: boolean;
	    showAuth: boolean;
	    toastPosition: string;
	    paneLocked: boolean;
	    sidebarWidth: number;
	    listWidth: number;
	    sendDelaySeconds: number;
	    flagHighlight: string;
	    showShortcutHints: boolean;
	    showAccountEmail: boolean;
	    alwaysLoadImages: boolean;
	    avatarSource: string;
	    avatarStyle: string;
	    multiSelectEnabled: boolean;
	    showSelectedCount: boolean;
	    sidebarIndentGuides: boolean;
	    rowTemplate: string;
	    rowShowAvatar: boolean;
	    rowShowSnippet: boolean;
	    previewLines: number;
	    uiScale: string;
	    messageFontSize: number;
	    showFlaggedCount: boolean;
	    flagColorSync: boolean;
	    showOfflineIndicator: boolean;
	    swipeEnabled: boolean;
	    swipeLeftAction: string;
	    swipeRightAction: string;
	    composeVimMode: boolean;
	    downloadIncludeAttachments: boolean;
	    appVimMode: boolean;
	    language: string;
	    lowPowerMode: boolean;
	    autoSyncIntervalSeconds: number;
	    defaultEditorMode: string;
	    composeAutocomplete: boolean;
	    composeChips: boolean;
	    updateCheckFrequency: string;
	
	    static createFrom(source: any = {}) {
	        return new UIPrefsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.accent = source["accent"];
	        this.density = source["density"];
	        this.showMailboxBadge = source["showMailboxBadge"];
	        this.showDateTime = source["showDateTime"];
	        this.showPgp = source["showPgp"];
	        this.showAuth = source["showAuth"];
	        this.toastPosition = source["toastPosition"];
	        this.paneLocked = source["paneLocked"];
	        this.sidebarWidth = source["sidebarWidth"];
	        this.listWidth = source["listWidth"];
	        this.sendDelaySeconds = source["sendDelaySeconds"];
	        this.flagHighlight = source["flagHighlight"];
	        this.showShortcutHints = source["showShortcutHints"];
	        this.showAccountEmail = source["showAccountEmail"];
	        this.alwaysLoadImages = source["alwaysLoadImages"];
	        this.avatarSource = source["avatarSource"];
	        this.avatarStyle = source["avatarStyle"];
	        this.multiSelectEnabled = source["multiSelectEnabled"];
	        this.showSelectedCount = source["showSelectedCount"];
	        this.sidebarIndentGuides = source["sidebarIndentGuides"];
	        this.rowTemplate = source["rowTemplate"];
	        this.rowShowAvatar = source["rowShowAvatar"];
	        this.rowShowSnippet = source["rowShowSnippet"];
	        this.previewLines = source["previewLines"];
	        this.uiScale = source["uiScale"];
	        this.messageFontSize = source["messageFontSize"];
	        this.showFlaggedCount = source["showFlaggedCount"];
	        this.flagColorSync = source["flagColorSync"];
	        this.showOfflineIndicator = source["showOfflineIndicator"];
	        this.swipeEnabled = source["swipeEnabled"];
	        this.swipeLeftAction = source["swipeLeftAction"];
	        this.swipeRightAction = source["swipeRightAction"];
	        this.composeVimMode = source["composeVimMode"];
	        this.downloadIncludeAttachments = source["downloadIncludeAttachments"];
	        this.appVimMode = source["appVimMode"];
	        this.language = source["language"];
	        this.lowPowerMode = source["lowPowerMode"];
	        this.autoSyncIntervalSeconds = source["autoSyncIntervalSeconds"];
	        this.defaultEditorMode = source["defaultEditorMode"];
	        this.composeAutocomplete = source["composeAutocomplete"];
	        this.composeChips = source["composeChips"];
	        this.updateCheckFrequency = source["updateCheckFrequency"];
	    }
	}
	export class UnifiedViewDTO {
	    key: string;
	    label: string;
	    unreadCount: number;
	    totalCount: number;
	
	    static createFrom(source: any = {}) {
	        return new UnifiedViewDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.unreadCount = source["unreadCount"];
	        this.totalCount = source["totalCount"];
	    }
	}
	export class UpdateCheckResult {
	    checked: boolean;
	    available: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    releaseUrl: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.checked = source["checked"];
	        this.available = source["available"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.releaseUrl = source["releaseUrl"];
	        this.error = source["error"];
	    }
	}

}

