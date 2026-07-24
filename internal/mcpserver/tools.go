package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Mailbox is the read-only view of Pelton's cached mail the tools operate on.
// The desktop layer implements it against the store and search index. Every
// method reads; none mutate mail. IDs are the local database ids the same
// backend hands the UI.
type Mailbox interface {
	// ListAccounts returns every configured account.
	ListAccounts(ctx context.Context) ([]Account, error)
	// ListFolders returns the folders of one account.
	ListFolders(ctx context.Context, accountID int64) ([]Folder, error)
	// ListMessages returns up to limit message summaries in a folder, newest
	// first. A non-positive limit applies a sensible default.
	ListMessages(ctx context.Context, folderID int64, limit int) ([]MessageSummary, error)
	// GetMessage returns one full message (headers, plain-text body and
	// attachment metadata) by id, or an error if it is not cached.
	GetMessage(ctx context.Context, id int64) (*Message, error)
	// Search runs a ranked full-text search over cached mail.
	Search(ctx context.Context, params SearchParams) ([]MessageSummary, error)
}

// Account is one configured mail account.
type Account struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name,omitempty"`
}

// Folder is one mailbox within an account.
type Folder struct {
	ID        int64  `json:"id"`
	AccountID int64  `json:"account_id"`
	Name      string `json:"name"`
	// Path is the raw IMAP mailbox name.
	Path string `json:"path"`
}

// MessageSummary is a message row without the body, for lists and search hits.
type MessageSummary struct {
	ID          int64  `json:"id"`
	AccountID   int64  `json:"account_id"`
	FolderID    int64  `json:"folder_id"`
	Subject     string `json:"subject"`
	FromName    string `json:"from_name,omitempty"`
	FromAddress string `json:"from_address"`
	// Date is RFC3339.
	Date           string `json:"date"`
	Unread         bool   `json:"unread"`
	Flagged        bool   `json:"flagged"`
	HasAttachments bool   `json:"has_attachments"`
}

// Attachment is attachment metadata only; bytes are never exposed over MCP.
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type,omitempty"`
	SizeBytes   int64  `json:"size_bytes"`
}

// Message is a full message: the summary fields plus recipients, the
// plain-text body and attachment metadata.
type Message struct {
	MessageSummary
	To          string       `json:"to,omitempty"`
	Cc          string       `json:"cc,omitempty"`
	BodyText    string       `json:"body_text"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// SearchParams mirrors the app's search: free text plus optional field scopes
// and a result cap. Empty fields are ignored.
type SearchParams struct {
	Query   string
	From    string
	To      string
	Subject string
	Limit   int
}

// tool input types. Field docs become the JSON-schema descriptions the agent
// sees, so they are written for that audience.

type listFoldersInput struct {
	AccountID int64 `json:"account_id" jsonschema:"the account id, from list_accounts"`
}

type listMessagesInput struct {
	FolderID int64 `json:"folder_id" jsonschema:"the folder id, from list_folders"`
	Limit    int   `json:"limit,omitempty" jsonschema:"max messages to return, newest first (default 50)"`
}

type getMessageInput struct {
	ID int64 `json:"id" jsonschema:"the message id, from list_messages or search_messages"`
}

type searchInput struct {
	Query   string `json:"query,omitempty" jsonschema:"free-text search over subject, sender, recipients and body"`
	From    string `json:"from,omitempty" jsonschema:"restrict to a sender name or address"`
	To      string `json:"to,omitempty" jsonschema:"restrict to a recipient"`
	Subject string `json:"subject,omitempty" jsonschema:"restrict to the subject line"`
	Limit   int    `json:"limit,omitempty" jsonschema:"max results (default 50)"`
}

// tool output wrappers, so structured results are objects.

type accountsOutput struct {
	Accounts []Account `json:"accounts"`
}

type foldersOutput struct {
	Folders []Folder `json:"folders"`
}

type messagesOutput struct {
	Messages []MessageSummary `json:"messages"`
}

// registerTools adds the read-only tool set to srv.
func registerTools(srv *mcp.Server, mb Mailbox) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_accounts",
		Description: "List the configured mail accounts.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, accountsOutput, error) {
		accts, err := mb.ListAccounts(ctx)
		if err != nil {
			return nil, accountsOutput{}, err
		}
		return nil, accountsOutput{Accounts: accts}, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_folders",
		Description: "List the folders (mailboxes) of one account.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listFoldersInput) (*mcp.CallToolResult, foldersOutput, error) {
		folders, err := mb.ListFolders(ctx, in.AccountID)
		if err != nil {
			return nil, foldersOutput{}, err
		}
		return nil, foldersOutput{Folders: folders}, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_messages",
		Description: "List messages in a folder, newest first. Returns summaries; use get_message for the body.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listMessagesInput) (*mcp.CallToolResult, messagesOutput, error) {
		msgs, err := mb.ListMessages(ctx, in.FolderID, in.Limit)
		if err != nil {
			return nil, messagesOutput{}, err
		}
		return nil, messagesOutput{Messages: msgs}, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_message",
		Description: "Get one full message: headers, plain-text body and attachment metadata (never attachment bytes).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getMessageInput) (*mcp.CallToolResult, Message, error) {
		msg, err := mb.GetMessage(ctx, in.ID)
		if err != nil {
			return nil, Message{}, err
		}
		return nil, *msg, nil
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search_messages",
		Description: "Ranked full-text search over cached mail. Combine free text with optional from/to/subject scopes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in searchInput) (*mcp.CallToolResult, messagesOutput, error) {
		hits, err := mb.Search(ctx, SearchParams{
			Query:   in.Query,
			From:    in.From,
			To:      in.To,
			Subject: in.Subject,
			Limit:   in.Limit,
		})
		if err != nil {
			return nil, messagesOutput{}, err
		}
		return nil, messagesOutput{Messages: hits}, nil
	})
}
