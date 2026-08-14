export namespace desktop {
	
	export class AccountDTO {
	    id: number;
	    email: string;
	    displayName: string;
	    username: string;
	    imapHost: string;
	    imapPort: number;
	    smtpHost: string;
	    smtpPort: number;
	    local: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AccountDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.email = source["email"];
	        this.displayName = source["displayName"];
	        this.username = source["username"];
	        this.imapHost = source["imapHost"];
	        this.imapPort = source["imapPort"];
	        this.smtpHost = source["smtpHost"];
	        this.smtpPort = source["smtpPort"];
	        this.local = source["local"];
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
	    username: string;
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
	        this.username = source["username"];
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
	export class VerdictDTO {
	    status: string;
	    malicious: number;
	    suspicious: number;
	    total: number;
	    permalink: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new VerdictDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.malicious = source["malicious"];
	        this.suspicious = source["suspicious"];
	        this.total = source["total"];
	        this.permalink = source["permalink"];
	        this.error = source["error"];
	    }
	}
	export class AttachmentVerdictDTO {
	    attachmentId: number;
	    filename: string;
	    verdict: VerdictDTO;
	
	    static createFrom(source: any = {}) {
	        return new AttachmentVerdictDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.attachmentId = source["attachmentId"];
	        this.filename = source["filename"];
	        this.verdict = this.convertValues(source["verdict"], VerdictDTO);
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
	export class BackupInfoDTO {
	    path: string;
	    createdAt: string;
	    appVersion: string;
	    hasSettings: boolean;
	    hasWhitelist: boolean;
	    hasMailboxes: boolean;
	    hasSignatures: boolean;
	    hasEncryptedCredentials: boolean;
	    settingCount: number;
	    mailboxCount: number;
	    signatureCount: number;
	
	    static createFrom(source: any = {}) {
	        return new BackupInfoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.createdAt = source["createdAt"];
	        this.appVersion = source["appVersion"];
	        this.hasSettings = source["hasSettings"];
	        this.hasWhitelist = source["hasWhitelist"];
	        this.hasMailboxes = source["hasMailboxes"];
	        this.hasSignatures = source["hasSignatures"];
	        this.hasEncryptedCredentials = source["hasEncryptedCredentials"];
	        this.settingCount = source["settingCount"];
	        this.mailboxCount = source["mailboxCount"];
	        this.signatureCount = source["signatureCount"];
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
	    sendAt: string;
	
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
	        this.sendAt = source["sendAt"];
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
	export class CreateFolderRequest {
	    accountId: number;
	    parentId: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateFolderRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.accountId = source["accountId"];
	        this.parentId = source["parentId"];
	        this.name = source["name"];
	    }
	}
	export class DefaultMailStatusDTO {
	    known: boolean;
	    isDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DefaultMailStatusDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.known = source["known"];
	        this.isDefault = source["isDefault"];
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
	export class FetchOlderResult {
	    fetched: number;
	    hasOlder: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FetchOlderResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fetched = source["fetched"];
	        this.hasOlder = source["hasOlder"];
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
	    pinned: boolean;
	    roleOverride: string;
	
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
	        this.pinned = source["pinned"];
	        this.roleOverride = source["roleOverride"];
	    }
	}
	export class ImageAllowEntryDTO {
	    value: string;
	    kind: string;
	    exampleMessageId: number;
	    exampleSubject: string;
	    exampleFrom: string;
	
	    static createFrom(source: any = {}) {
	        return new ImageAllowEntryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.kind = source["kind"];
	        this.exampleMessageId = source["exampleMessageId"];
	        this.exampleSubject = source["exampleSubject"];
	        this.exampleFrom = source["exampleFrom"];
	    }
	}
	export class LinkVerdictDTO {
	    url: string;
	    verdict: VerdictDTO;
	
	    static createFrom(source: any = {}) {
	        return new LinkVerdictDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.verdict = this.convertValues(source["verdict"], VerdictDTO);
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
	export class ListMessagesRequest {
	    kind: string;
	    folderId: number;
	    view: string;
	    viewId: number;
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
	        this.viewId = source["viewId"];
	        this.limit = source["limit"];
	        this.offset = source["offset"];
	    }
	}
	export class LogStatusDTO {
	    dir: string;
	    writing: boolean;
	    forced: boolean;
	    sizeBytes: number;
	    crashName: string;
	    crashTime: string;
	
	    static createFrom(source: any = {}) {
	        return new LogStatusDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.writing = source["writing"];
	        this.forced = source["forced"];
	        this.sizeBytes = source["sizeBytes"];
	        this.crashName = source["crashName"];
	        this.crashTime = source["crashTime"];
	    }
	}
	export class MCPConfigDTO {
	    enabled: boolean;
	    port: number;
	    token: string;
	    url: string;
	    running: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MCPConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.port = source["port"];
	        this.token = source["token"];
	        this.url = source["url"];
	        this.running = source["running"];
	    }
	}
	export class MailtoDraft {
	    to: string;
	    cc: string;
	    bcc: string;
	    subject: string;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new MailtoDraft(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.to = source["to"];
	        this.cc = source["cc"];
	        this.bcc = source["bcc"];
	        this.subject = source["subject"];
	        this.body = source["body"];
	    }
	}
	export class UnsubscribeDTO {
	    kind: string;
	    target: string;
	    done: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UnsubscribeDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.target = source["target"];
	        this.done = source["done"];
	    }
	}
	export class TrackingPixelDTO {
	    host: string;
	    url: string;
	    reasons: string[];
	
	    static createFrom(source: any = {}) {
	        return new TrackingPixelDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.url = source["url"];
	        this.reasons = source["reasons"];
	    }
	}
	export class SMIMEDTO {
	    status: string;
	    signer: string;
	    email: string;
	    issuer: string;
	    detail: string;
	
	    static createFrom(source: any = {}) {
	        return new SMIMEDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.signer = source["signer"];
	        this.email = source["email"];
	        this.issuer = source["issuer"];
	        this.detail = source["detail"];
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
	    senderVip: boolean;
	    smime: SMIMEDTO;
	    toAddresses: string;
	    ccAddresses: string;
	    bodyPlain: string;
	    bodyHtmlSafe: string;
	    bodyQuote: string;
	    isHtml: boolean;
	    hasRemoteContent: boolean;
	    remoteAllowed: boolean;
	    remoteHosts: string[];
	    trackingPixels: TrackingPixelDTO[];
	    attachments: AttachmentDTO[];
	    pgpState: string;
	    unsubscribe?: UnsubscribeDTO;
	
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
	        this.senderVip = source["senderVip"];
	        this.smime = this.convertValues(source["smime"], SMIMEDTO);
	        this.toAddresses = source["toAddresses"];
	        this.ccAddresses = source["ccAddresses"];
	        this.bodyPlain = source["bodyPlain"];
	        this.bodyHtmlSafe = source["bodyHtmlSafe"];
	        this.bodyQuote = source["bodyQuote"];
	        this.isHtml = source["isHtml"];
	        this.hasRemoteContent = source["hasRemoteContent"];
	        this.remoteAllowed = source["remoteAllowed"];
	        this.remoteHosts = source["remoteHosts"];
	        this.trackingPixels = this.convertValues(source["trackingPixels"], TrackingPixelDTO);
	        this.attachments = this.convertValues(source["attachments"], AttachmentDTO);
	        this.pgpState = source["pgpState"];
	        this.unsubscribe = this.convertValues(source["unsubscribe"], UnsubscribeDTO);
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
	    senderVip: boolean;
	    smime: SMIMEDTO;
	
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
	        this.senderVip = source["senderVip"];
	        this.smime = this.convertValues(source["smime"], SMIMEDTO);
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
	export class MessageListDTO {
	    messages: MessageSummaryDTO[];
	    total: number;
	    hasOlder: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MessageListDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.messages = this.convertValues(source["messages"], MessageSummaryDTO);
	        this.total = source["total"];
	        this.hasOlder = source["hasOlder"];
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
	export class MessageScanDTO {
	    links: LinkVerdictDTO[];
	    attachments: AttachmentVerdictDTO[];
	
	    static createFrom(source: any = {}) {
	        return new MessageScanDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.links = this.convertValues(source["links"], LinkVerdictDTO);
	        this.attachments = this.convertValues(source["attachments"], AttachmentVerdictDTO);
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
	export class PGPKeyDTO {
	    fingerprint: string;
	    name: string;
	    email: string;
	    emails: string[];
	    created: string;
	    expires: string;
	    expired: boolean;
	    hasPrivate: boolean;
	    locked: boolean;
	    unlocked: boolean;
	    remembered: boolean;
	    algorithm: string;
	    bits: number;
	
	    static createFrom(source: any = {}) {
	        return new PGPKeyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fingerprint = source["fingerprint"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.emails = source["emails"];
	        this.created = source["created"];
	        this.expires = source["expires"];
	        this.expired = source["expired"];
	        this.hasPrivate = source["hasPrivate"];
	        this.locked = source["locked"];
	        this.unlocked = source["unlocked"];
	        this.remembered = source["remembered"];
	        this.algorithm = source["algorithm"];
	        this.bits = source["bits"];
	    }
	}
	export class PendingMailtoDTO {
	    present: boolean;
	    draft: MailtoDraft;
	
	    static createFrom(source: any = {}) {
	        return new PendingMailtoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.present = source["present"];
	        this.draft = this.convertValues(source["draft"], MailtoDraft);
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
	export class ProxyConfigDTO {
	    mode: string;
	    scheme: string;
	    host: string;
	    port: number;
	    username: string;
	    password: string;
	    hasPassword: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProxyConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.scheme = source["scheme"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.hasPassword = source["hasPassword"];
	    }
	}
	
	export class SaveThemeRequest {
	    id: string;
	    name: string;
	    author: string;
	    version: string;
	    base: string;
	    tokens: Record<string, string>;
	    css: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveThemeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.author = source["author"];
	        this.version = source["version"];
	        this.base = source["base"];
	        this.tokens = source["tokens"];
	        this.css = source["css"];
	    }
	}
	export class SearchRequestDTO {
	    query: string;
	    afterUnix: number;
	    beforeUnix: number;
	    limit: number;
	    offset: number;
	    from: string;
	    to: string;
	    subject: string;
	    hasAttachment: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SearchRequestDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.afterUnix = source["afterUnix"];
	        this.beforeUnix = source["beforeUnix"];
	        this.limit = source["limit"];
	        this.offset = source["offset"];
	        this.from = source["from"];
	        this.to = source["to"];
	        this.subject = source["subject"];
	        this.hasAttachment = source["hasAttachment"];
	    }
	}
	export class SearchResultDTO {
	    messages: MessageSummaryDTO[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchResultDTO(source);
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
	    username: string;
	    imapHost: string;
	    imapPort: number;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new TestConnectionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.email = source["email"];
	        this.username = source["username"];
	        this.imapHost = source["imapHost"];
	        this.imapPort = source["imapPort"];
	        this.password = source["password"];
	    }
	}
	export class ThemeApplyDTO {
	    id: string;
	    base: string;
	    tokens: Record<string, string>;
	    css: string;
	    icons: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ThemeApplyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.base = source["base"];
	        this.tokens = source["tokens"];
	        this.css = source["css"];
	        this.icons = source["icons"];
	    }
	}
	export class ThemeDraftDTO {
	    id: string;
	    name: string;
	    author: string;
	    version: string;
	    base: string;
	    tokens: Record<string, string>;
	    css: string;
	
	    static createFrom(source: any = {}) {
	        return new ThemeDraftDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.author = source["author"];
	        this.version = source["version"];
	        this.base = source["base"];
	        this.tokens = source["tokens"];
	        this.css = source["css"];
	    }
	}
	export class ThemeInfoDTO {
	    id: string;
	    name: string;
	    author: string;
	    version: string;
	    description: string;
	    base: string;
	    hasCss: boolean;
	    remoteRefs: string[];
	    preview: string;
	    compatWarning: string;
	    swatches: string[];
	
	    static createFrom(source: any = {}) {
	        return new ThemeInfoDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.author = source["author"];
	        this.version = source["version"];
	        this.description = source["description"];
	        this.base = source["base"];
	        this.hasCss = source["hasCss"];
	        this.remoteRefs = source["remoteRefs"];
	        this.preview = source["preview"];
	        this.compatWarning = source["compatWarning"];
	        this.swatches = source["swatches"];
	    }
	}
	export class ThemeImportPreviewDTO {
	    canceled: boolean;
	    path: string;
	    info: ThemeInfoDTO;
	    cssFiles: themepack.CSSFile[];
	    tokenCount: number;
	    updatesExisting: boolean;
	    installedVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new ThemeImportPreviewDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.canceled = source["canceled"];
	        this.path = source["path"];
	        this.info = this.convertValues(source["info"], ThemeInfoDTO);
	        this.cssFiles = this.convertValues(source["cssFiles"], themepack.CSSFile);
	        this.tokenCount = source["tokenCount"];
	        this.updatesExisting = source["updatesExisting"];
	        this.installedVersion = source["installedVersion"];
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
	
	export class ThunderbirdAccountDTO {
	    email: string;
	    displayName: string;
	    username: string;
	    imapHost: string;
	    imapPort: number;
	    smtpHost: string;
	    smtpPort: number;
	    kind: string;
	    exists: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ThunderbirdAccountDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.email = source["email"];
	        this.displayName = source["displayName"];
	        this.username = source["username"];
	        this.imapHost = source["imapHost"];
	        this.imapPort = source["imapPort"];
	        this.smtpHost = source["smtpHost"];
	        this.smtpPort = source["smtpPort"];
	        this.kind = source["kind"];
	        this.exists = source["exists"];
	    }
	}
	export class ThunderbirdFolderDTO {
	    name: string;
	    path: string;
	    sizeBytes: number;
	
	    static createFrom(source: any = {}) {
	        return new ThunderbirdFolderDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.sizeBytes = source["sizeBytes"];
	    }
	}
	export class ThunderbirdProfileDTO {
	    name: string;
	    path: string;
	    accounts: ThunderbirdAccountDTO[];
	    localFolders: ThunderbirdFolderDTO[];
	
	    static createFrom(source: any = {}) {
	        return new ThunderbirdProfileDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.accounts = this.convertValues(source["accounts"], ThunderbirdAccountDTO);
	        this.localFolders = this.convertValues(source["localFolders"], ThunderbirdFolderDTO);
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
	    blockTrackingPixels: boolean;
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
	    viewsPlacement: string;
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
	    emptyStateImage: string;
	    emptyStateFullscreen: boolean;
	    cornerStyle: string;
	    themeId: string;
	    menuBarInApp: boolean;
	    menuBarNativeMinimal: boolean;
	    menuBarIcons: boolean;
	    timeFormat: string;
	    reduceMotion: boolean;
	    handCursor: boolean;
	    dockBadge: boolean;
	    themeDarkStart: string;
	    themeDarkEnd: string;
	    bodyFont: string;
	    uiFont: string;
	    monoFont: string;
	    notifyNewMail: boolean;
	    verboseSync: boolean;
	    closeAction: string;
	    syncMessageLimit: number;
	    syncAutoBackfill: boolean;
	    startupSelection: string;
	    logToFile: boolean;
	    logLevel: string;
	    logMessageMetadata: boolean;
	    crashLogs: boolean;
	
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
	        this.blockTrackingPixels = source["blockTrackingPixels"];
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
	        this.viewsPlacement = source["viewsPlacement"];
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
	        this.emptyStateImage = source["emptyStateImage"];
	        this.emptyStateFullscreen = source["emptyStateFullscreen"];
	        this.cornerStyle = source["cornerStyle"];
	        this.themeId = source["themeId"];
	        this.menuBarInApp = source["menuBarInApp"];
	        this.menuBarNativeMinimal = source["menuBarNativeMinimal"];
	        this.menuBarIcons = source["menuBarIcons"];
	        this.timeFormat = source["timeFormat"];
	        this.reduceMotion = source["reduceMotion"];
	        this.handCursor = source["handCursor"];
	        this.dockBadge = source["dockBadge"];
	        this.themeDarkStart = source["themeDarkStart"];
	        this.themeDarkEnd = source["themeDarkEnd"];
	        this.bodyFont = source["bodyFont"];
	        this.uiFont = source["uiFont"];
	        this.monoFont = source["monoFont"];
	        this.notifyNewMail = source["notifyNewMail"];
	        this.verboseSync = source["verboseSync"];
	        this.closeAction = source["closeAction"];
	        this.syncMessageLimit = source["syncMessageLimit"];
	        this.syncAutoBackfill = source["syncAutoBackfill"];
	        this.startupSelection = source["startupSelection"];
	        this.logToFile = source["logToFile"];
	        this.logLevel = source["logLevel"];
	        this.logMessageMetadata = source["logMessageMetadata"];
	        this.crashLogs = source["crashLogs"];
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
	
	export class UpdateAccountRequest {
	    id: number;
	    displayName: string;
	    username: string;
	    imapHost: string;
	    imapPort: number;
	    smtpHost: string;
	    smtpPort: number;
	    password: string;

	    static createFrom(source: any = {}) {
	        return new UpdateAccountRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.displayName = source["displayName"];
	        this.username = source["username"];
	        this.imapHost = source["imapHost"];
	        this.imapPort = source["imapPort"];
	        this.smtpHost = source["smtpHost"];
	        this.smtpPort = source["smtpPort"];
	        this.password = source["password"];
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
	export class UserLocaleApplyDTO {
	    id: string;
	    name: string;
	    base: string;
	    strings: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new UserLocaleApplyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.base = source["base"];
	        this.strings = source["strings"];
	    }
	}
	export class UserLocaleDTO {
	    id: string;
	    name: string;
	    author: string;
	    base: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new UserLocaleDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.author = source["author"];
	        this.base = source["base"];
	        this.count = source["count"];
	    }
	}
	
	export class ViewDTO {
	    id: number;
	    name: string;
	    icon: string;
	    color: string;
	    queryText: string;
	    queryFrom: string[];
	    queryTo: string[];
	    querySubject: string;
	    withinDays: number;
	    useRegex: boolean;
	    unreadOnly: boolean;
	    flaggedOnly: boolean;
	    hasAttachment: boolean;
	    accountId: number;
	    position: number;
	    unreadCount: number;
	    totalCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ViewDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.queryText = source["queryText"];
	        this.queryFrom = source["queryFrom"];
	        this.queryTo = source["queryTo"];
	        this.querySubject = source["querySubject"];
	        this.withinDays = source["withinDays"];
	        this.useRegex = source["useRegex"];
	        this.unreadOnly = source["unreadOnly"];
	        this.flaggedOnly = source["flaggedOnly"];
	        this.hasAttachment = source["hasAttachment"];
	        this.accountId = source["accountId"];
	        this.position = source["position"];
	        this.unreadCount = source["unreadCount"];
	        this.totalCount = source["totalCount"];
	    }
	}
	export class VirusTotalConfigDTO {
	    enabled: boolean;
	    hasApiKey: boolean;
	    autoScanLinks: boolean;
	    autoScanAttachments: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VirusTotalConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.hasApiKey = source["hasApiKey"];
	        this.autoScanLinks = source["autoScanLinks"];
	        this.autoScanAttachments = source["autoScanAttachments"];
	    }
	}

}

export namespace themepack {
	
	export class CSSFile {
	    path: string;
	    content: string;
	    remoteRefs: string[];
	
	    static createFrom(source: any = {}) {
	        return new CSSFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.content = source["content"];
	        this.remoteRefs = source["remoteRefs"];
	    }
	}

}

