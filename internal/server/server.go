// Package server wires the Mochi Cards API to MCP tools exposed over the
// Model Context Protocol.
package server

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/dudleycarr/mochi-card-mcp/internal/mochi"
)

// cardAPI is the subset of the Mochi client used by the tool handlers. It is an
// interface so the handlers can be tested without a live API.
type cardAPI interface {
	ListCards(ctx context.Context, params mochi.ListCardsParams) (mochi.CardsResult, error)
	ListDueCards(ctx context.Context, params mochi.DueCardsParams) ([]mochi.Card, error)
	GetCard(ctx context.Context, id string) (mochi.Card, error)
	CreateCard(ctx context.Context, params mochi.CreateCardParams) (mochi.Card, error)
	UpdateCard(ctx context.Context, id string, params mochi.UpdateCardParams) (mochi.Card, error)
	DeleteCard(ctx context.Context, id string) error
	SearchCards(ctx context.Context, query, bookmark string) (mochi.CardsResult, error)
	ListDecks(ctx context.Context, bookmark string) (mochi.DecksResult, error)
	GetDeck(ctx context.Context, id string) (mochi.Deck, error)
	CreateDeck(ctx context.Context, params mochi.CreateDeckParams) (mochi.Deck, error)
	UpdateDeck(ctx context.Context, id string, params mochi.UpdateDeckParams) (mochi.Deck, error)
	DeleteDeck(ctx context.Context, id string) error
	ListTemplates(ctx context.Context, bookmark string) (mochi.TemplatesResult, error)
	GetTemplate(ctx context.Context, id string) (mochi.Template, error)
	CreateTemplate(ctx context.Context, params mochi.CreateTemplateParams) (mochi.Template, error)
	AddCardAttachment(ctx context.Context, cardID, filename string, data []byte, contentType string) (mochi.Card, error)
	DeleteCardAttachment(ctx context.Context, cardID, filename string) error
}

// handlers holds the dependencies shared by the tool handlers.
type handlers struct {
	api cardAPI
}

// New creates an MCP server with all Mochi tools registered, backed by the
// given client.
func New(client *mochi.Client, version string) *mcp.Server {
	return newServer(client, version)
}

func newServer(api cardAPI, version string) *mcp.Server {
	h := &handlers{api: api}
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "mochi-card-mcp",
		Title:   "Mochi Cards",
		Version: version,
	}, nil)
	h.register(s)
	return s
}

func (h *handlers) register(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_list_cards",
		Description: "List Mochi flashcards, optionally filtered by deck. Supports pagination via bookmark.",
	}, h.listCards)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_list_due_cards",
		Description: "List Mochi flashcards that are due for review, optionally filtered by deck and/or a specific date.",
	}, h.listDueCards)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_get_card",
		Description: "Get a single Mochi flashcard by its ID.",
	}, h.getCard)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_create_card",
		Description: "Create a Mochi flashcard. 'name' is the front, 'content' is the back; both are Markdown.",
	}, h.createCard)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_update_card",
		Description: "Update a Mochi flashcard's front ('name') and/or back ('content').",
	}, h.updateCard)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_delete_card",
		Description: "Permanently delete a Mochi flashcard by its ID.",
	}, h.deleteCard)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_search_cards",
		Description: "Search Mochi flashcards whose content contains the query (case-insensitive). Supports pagination via bookmark.",
	}, h.searchCards)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_list_decks",
		Description: "List Mochi decks. Supports pagination via bookmark.",
	}, h.listDecks)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_get_deck",
		Description: "Get a single Mochi deck by its ID.",
	}, h.getDeck)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_create_deck",
		Description: "Create a Mochi deck with the given name and optional parent deck.",
	}, h.createDeck)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_update_deck",
		Description: "Update a Mochi deck's name, parent deck, and/or sort position.",
	}, h.updateDeck)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_delete_deck",
		Description: "Permanently delete a Mochi deck by its ID.",
	}, h.deleteDeck)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_list_templates",
		Description: "List Mochi card templates. Supports pagination via bookmark.",
	}, h.listTemplates)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_get_template",
		Description: "Get a single Mochi card template by its ID.",
	}, h.getTemplate)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_create_template",
		Description: "Create a Mochi card template. Content is Markdown with << Field name >> placeholders; fields maps each field ID to its definition.",
	}, h.createTemplate)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_add_card_attachment",
		Description: "Attach a file to a Mochi card. Provide the file as base64; reference it from card content as ![](@media/<filename>).",
	}, h.addCardAttachment)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "mochi_delete_card_attachment",
		Description: "Delete an attachment from a Mochi card by filename.",
	}, h.deleteCardAttachment)
}

// ---- Input/output types ----

// ListCardsInput are the arguments for mochi_list_cards.
type ListCardsInput struct {
	DeckID   string `json:"deck_id,omitempty" jsonschema:"optional deck ID to filter cards by"`
	Bookmark string `json:"bookmark,omitempty" jsonschema:"pagination bookmark from a previous response"`
	Limit    int    `json:"limit,omitempty" jsonschema:"maximum number of cards to return (1-100)"`
}

// CardsOutput is a page of cards returned by list and search tools.
type CardsOutput struct {
	Cards    []mochi.Card `json:"cards" jsonschema:"the cards on this page"`
	Bookmark string       `json:"bookmark,omitempty" jsonschema:"bookmark to fetch the next page, if any"`
}

// ListDueCardsInput are the arguments for mochi_list_due_cards.
type ListDueCardsInput struct {
	DeckID string `json:"deck_id,omitempty" jsonschema:"optional deck ID to restrict due cards to"`
	Date   string `json:"date,omitempty" jsonschema:"optional date (timestamp) the cards are due on; defaults to today"`
}

// DueCardsOutput is the set of due cards. The due endpoint is not paginated, so
// there is no bookmark.
type DueCardsOutput struct {
	Cards []mochi.Card `json:"cards" jsonschema:"the due cards"`
}

// GetCardInput are the arguments for mochi_get_card.
type GetCardInput struct {
	CardID string `json:"card_id" jsonschema:"the ID of the card to fetch"`
}

// CardOutput wraps a single card.
type CardOutput struct {
	Card mochi.Card `json:"card" jsonschema:"the card"`
}

// CreateCardInput are the arguments for mochi_create_card.
type CreateCardInput struct {
	Name    string `json:"name" jsonschema:"the front of the card (Markdown)"`
	Content string `json:"content" jsonschema:"the back of the card (Markdown)"`
	DeckID  string `json:"deck_id,omitempty" jsonschema:"optional deck ID to add the card to"`
}

// UpdateCardInput are the arguments for mochi_update_card.
type UpdateCardInput struct {
	CardID  string `json:"card_id" jsonschema:"the ID of the card to update"`
	Name    string `json:"name,omitempty" jsonschema:"new front of the card (Markdown); leave empty to keep the current front"`
	Content string `json:"content,omitempty" jsonschema:"new back of the card (Markdown); leave empty to keep the current back"`
}

// DeleteCardInput are the arguments for mochi_delete_card.
type DeleteCardInput struct {
	CardID string `json:"card_id" jsonschema:"the ID of the card to delete"`
}

// DeleteCardOutput reports the result of a delete.
type DeleteCardOutput struct {
	Deleted bool   `json:"deleted" jsonschema:"true if the card was deleted"`
	CardID  string `json:"card_id" jsonschema:"the ID of the deleted card"`
}

// SearchCardsInput are the arguments for mochi_search_cards.
type SearchCardsInput struct {
	Query    string `json:"query" jsonschema:"text to search for within card content"`
	Bookmark string `json:"bookmark,omitempty" jsonschema:"pagination bookmark from a previous response"`
}

// ListDecksInput are the arguments for mochi_list_decks.
type ListDecksInput struct {
	Bookmark string `json:"bookmark,omitempty" jsonschema:"pagination bookmark from a previous response"`
}

// DecksOutput is a page of decks.
type DecksOutput struct {
	Decks    []mochi.Deck `json:"decks" jsonschema:"the decks on this page"`
	Bookmark string       `json:"bookmark,omitempty" jsonschema:"bookmark to fetch the next page, if any"`
}

// GetDeckInput are the arguments for mochi_get_deck.
type GetDeckInput struct {
	DeckID string `json:"deck_id" jsonschema:"the ID of the deck to fetch"`
}

// CreateDeckInput are the arguments for mochi_create_deck.
type CreateDeckInput struct {
	Name     string `json:"name" jsonschema:"the name of the deck"`
	ParentID string `json:"parent_id,omitempty" jsonschema:"optional parent deck ID"`
}

// UpdateDeckInput are the arguments for mochi_update_deck. Leave a field empty
// (or sort null) to keep its current value.
type UpdateDeckInput struct {
	DeckID   string `json:"deck_id" jsonschema:"the ID of the deck to update"`
	Name     string `json:"name,omitempty" jsonschema:"new deck name"`
	ParentID string `json:"parent_id,omitempty" jsonschema:"new parent deck ID"`
	Sort     *int   `json:"sort,omitempty" jsonschema:"new sort position among sibling decks"`
}

// DeleteDeckInput are the arguments for mochi_delete_deck.
type DeleteDeckInput struct {
	DeckID string `json:"deck_id" jsonschema:"the ID of the deck to delete"`
}

// DeleteDeckOutput reports the result of a deck delete.
type DeleteDeckOutput struct {
	Deleted bool   `json:"deleted" jsonschema:"true if the deck was deleted"`
	DeckID  string `json:"deck_id" jsonschema:"the ID of the deleted deck"`
}

// ListTemplatesInput are the arguments for mochi_list_templates.
type ListTemplatesInput struct {
	Bookmark string `json:"bookmark,omitempty" jsonschema:"pagination bookmark from a previous response"`
}

// TemplatesOutput is a page of templates.
type TemplatesOutput struct {
	Templates []mochi.Template `json:"templates" jsonschema:"the templates on this page"`
	Bookmark  string           `json:"bookmark,omitempty" jsonschema:"bookmark to fetch the next page, if any"`
}

// GetTemplateInput are the arguments for mochi_get_template.
type GetTemplateInput struct {
	TemplateID string `json:"template_id" jsonschema:"the ID of the template to fetch"`
}

// CreateTemplateInput are the arguments for mochi_create_template.
type CreateTemplateInput struct {
	Name    string                         `json:"name" jsonschema:"the template name"`
	Content string                         `json:"content" jsonschema:"the template body (Markdown with << Field name >> placeholders)"`
	Fields  map[string]mochi.TemplateField `json:"fields,omitempty" jsonschema:"map of field ID to its definition (id, name, type, pos)"`
}

// TemplateOutput wraps a single template.
type TemplateOutput struct {
	Template mochi.Template `json:"template" jsonschema:"the template"`
}

// AddCardAttachmentInput are the arguments for mochi_add_card_attachment.
type AddCardAttachmentInput struct {
	CardID      string `json:"card_id" jsonschema:"the ID of the card to attach the file to"`
	Filename    string `json:"filename" jsonschema:"the attachment filename, e.g. diagram.png"`
	DataBase64  string `json:"data_base64" jsonschema:"the file contents, base64-encoded"`
	ContentType string `json:"content_type,omitempty" jsonschema:"optional MIME type, e.g. image/png"`
}

// AddCardAttachmentOutput reports the result of an attachment upload.
type AddCardAttachmentOutput struct {
	CardID            string `json:"card_id" jsonschema:"the card the file was attached to"`
	Filename          string `json:"filename" jsonschema:"the attachment filename"`
	MarkdownReference string `json:"markdown_reference" jsonschema:"snippet to embed the attachment in card content"`
}

// DeleteCardAttachmentInput are the arguments for mochi_delete_card_attachment.
type DeleteCardAttachmentInput struct {
	CardID   string `json:"card_id" jsonschema:"the ID of the card"`
	Filename string `json:"filename" jsonschema:"the attachment filename to delete"`
}

// DeleteCardAttachmentOutput reports the result of an attachment delete.
type DeleteCardAttachmentOutput struct {
	Deleted  bool   `json:"deleted" jsonschema:"true if the attachment was deleted"`
	CardID   string `json:"card_id" jsonschema:"the card the attachment was removed from"`
	Filename string `json:"filename" jsonschema:"the deleted attachment filename"`
}

// DeckOutput wraps a single deck.
type DeckOutput struct {
	Deck mochi.Deck `json:"deck" jsonschema:"the deck"`
}

// ---- Handlers ----

func (h *handlers) listCards(ctx context.Context, _ *mcp.CallToolRequest, in ListCardsInput) (*mcp.CallToolResult, CardsOutput, error) {
	res, err := h.api.ListCards(ctx, mochi.ListCardsParams{
		DeckID:   in.DeckID,
		Bookmark: in.Bookmark,
		Limit:    in.Limit,
	})
	if err != nil {
		return nil, CardsOutput{}, err
	}
	return nil, CardsOutput{Cards: res.Docs, Bookmark: res.Bookmark}, nil
}

func (h *handlers) listDueCards(ctx context.Context, _ *mcp.CallToolRequest, in ListDueCardsInput) (*mcp.CallToolResult, DueCardsOutput, error) {
	cards, err := h.api.ListDueCards(ctx, mochi.DueCardsParams{
		DeckID: in.DeckID,
		Date:   in.Date,
	})
	if err != nil {
		return nil, DueCardsOutput{}, err
	}
	return nil, DueCardsOutput{Cards: cards}, nil
}

func (h *handlers) getCard(ctx context.Context, _ *mcp.CallToolRequest, in GetCardInput) (*mcp.CallToolResult, CardOutput, error) {
	if in.CardID == "" {
		return nil, CardOutput{}, fmt.Errorf("card_id is required")
	}
	card, err := h.api.GetCard(ctx, in.CardID)
	if err != nil {
		return nil, CardOutput{}, err
	}
	return nil, CardOutput{Card: card}, nil
}

func (h *handlers) createCard(ctx context.Context, _ *mcp.CallToolRequest, in CreateCardInput) (*mcp.CallToolResult, CardOutput, error) {
	if in.Name == "" {
		return nil, CardOutput{}, fmt.Errorf("name is required")
	}
	card, err := h.api.CreateCard(ctx, mochi.CreateCardParams{
		Content: mochi.JoinSides(in.Name, in.Content),
		DeckID:  in.DeckID,
	})
	if err != nil {
		return nil, CardOutput{}, err
	}
	return nil, CardOutput{Card: card}, nil
}

func (h *handlers) updateCard(ctx context.Context, _ *mcp.CallToolRequest, in UpdateCardInput) (*mcp.CallToolResult, CardOutput, error) {
	if in.CardID == "" {
		return nil, CardOutput{}, fmt.Errorf("card_id is required")
	}
	if in.Name == "" && in.Content == "" {
		return nil, CardOutput{}, fmt.Errorf("at least one of name or content is required")
	}

	// Merge with the existing card so a caller can update only one side.
	current, err := h.api.GetCard(ctx, in.CardID)
	if err != nil {
		return nil, CardOutput{}, err
	}
	front, back := mochi.SplitSides(current.Content)
	if in.Name != "" {
		front = in.Name
	}
	if in.Content != "" {
		back = in.Content
	}

	card, err := h.api.UpdateCard(ctx, in.CardID, mochi.UpdateCardParams{
		Content: mochi.JoinSides(front, back),
	})
	if err != nil {
		return nil, CardOutput{}, err
	}
	return nil, CardOutput{Card: card}, nil
}

func (h *handlers) deleteCard(ctx context.Context, _ *mcp.CallToolRequest, in DeleteCardInput) (*mcp.CallToolResult, DeleteCardOutput, error) {
	if in.CardID == "" {
		return nil, DeleteCardOutput{}, fmt.Errorf("card_id is required")
	}
	if err := h.api.DeleteCard(ctx, in.CardID); err != nil {
		return nil, DeleteCardOutput{}, err
	}
	return nil, DeleteCardOutput{Deleted: true, CardID: in.CardID}, nil
}

func (h *handlers) searchCards(ctx context.Context, _ *mcp.CallToolRequest, in SearchCardsInput) (*mcp.CallToolResult, CardsOutput, error) {
	if in.Query == "" {
		return nil, CardsOutput{}, fmt.Errorf("query is required")
	}
	res, err := h.api.SearchCards(ctx, in.Query, in.Bookmark)
	if err != nil {
		return nil, CardsOutput{}, err
	}
	return nil, CardsOutput{Cards: res.Docs, Bookmark: res.Bookmark}, nil
}

func (h *handlers) listDecks(ctx context.Context, _ *mcp.CallToolRequest, in ListDecksInput) (*mcp.CallToolResult, DecksOutput, error) {
	res, err := h.api.ListDecks(ctx, in.Bookmark)
	if err != nil {
		return nil, DecksOutput{}, err
	}
	return nil, DecksOutput{Decks: res.Docs, Bookmark: res.Bookmark}, nil
}

func (h *handlers) getDeck(ctx context.Context, _ *mcp.CallToolRequest, in GetDeckInput) (*mcp.CallToolResult, DeckOutput, error) {
	if in.DeckID == "" {
		return nil, DeckOutput{}, fmt.Errorf("deck_id is required")
	}
	deck, err := h.api.GetDeck(ctx, in.DeckID)
	if err != nil {
		return nil, DeckOutput{}, err
	}
	return nil, DeckOutput{Deck: deck}, nil
}

func (h *handlers) updateDeck(ctx context.Context, _ *mcp.CallToolRequest, in UpdateDeckInput) (*mcp.CallToolResult, DeckOutput, error) {
	if in.DeckID == "" {
		return nil, DeckOutput{}, fmt.Errorf("deck_id is required")
	}
	if in.Name == "" && in.ParentID == "" && in.Sort == nil {
		return nil, DeckOutput{}, fmt.Errorf("at least one of name, parent_id, or sort is required")
	}
	deck, err := h.api.UpdateDeck(ctx, in.DeckID, mochi.UpdateDeckParams{
		Name:     in.Name,
		ParentID: in.ParentID,
		Sort:     in.Sort,
	})
	if err != nil {
		return nil, DeckOutput{}, err
	}
	return nil, DeckOutput{Deck: deck}, nil
}

func (h *handlers) deleteDeck(ctx context.Context, _ *mcp.CallToolRequest, in DeleteDeckInput) (*mcp.CallToolResult, DeleteDeckOutput, error) {
	if in.DeckID == "" {
		return nil, DeleteDeckOutput{}, fmt.Errorf("deck_id is required")
	}
	if err := h.api.DeleteDeck(ctx, in.DeckID); err != nil {
		return nil, DeleteDeckOutput{}, err
	}
	return nil, DeleteDeckOutput{Deleted: true, DeckID: in.DeckID}, nil
}

func (h *handlers) listTemplates(ctx context.Context, _ *mcp.CallToolRequest, in ListTemplatesInput) (*mcp.CallToolResult, TemplatesOutput, error) {
	res, err := h.api.ListTemplates(ctx, in.Bookmark)
	if err != nil {
		return nil, TemplatesOutput{}, err
	}
	return nil, TemplatesOutput{Templates: res.Docs, Bookmark: res.Bookmark}, nil
}

func (h *handlers) getTemplate(ctx context.Context, _ *mcp.CallToolRequest, in GetTemplateInput) (*mcp.CallToolResult, TemplateOutput, error) {
	if in.TemplateID == "" {
		return nil, TemplateOutput{}, fmt.Errorf("template_id is required")
	}
	tmpl, err := h.api.GetTemplate(ctx, in.TemplateID)
	if err != nil {
		return nil, TemplateOutput{}, err
	}
	return nil, TemplateOutput{Template: tmpl}, nil
}

func (h *handlers) createTemplate(ctx context.Context, _ *mcp.CallToolRequest, in CreateTemplateInput) (*mcp.CallToolResult, TemplateOutput, error) {
	if in.Name == "" {
		return nil, TemplateOutput{}, fmt.Errorf("name is required")
	}
	if in.Content == "" {
		return nil, TemplateOutput{}, fmt.Errorf("content is required")
	}
	tmpl, err := h.api.CreateTemplate(ctx, mochi.CreateTemplateParams{
		Name:    in.Name,
		Content: in.Content,
		Fields:  in.Fields,
	})
	if err != nil {
		return nil, TemplateOutput{}, err
	}
	return nil, TemplateOutput{Template: tmpl}, nil
}

func (h *handlers) addCardAttachment(ctx context.Context, _ *mcp.CallToolRequest, in AddCardAttachmentInput) (*mcp.CallToolResult, AddCardAttachmentOutput, error) {
	if in.CardID == "" {
		return nil, AddCardAttachmentOutput{}, fmt.Errorf("card_id is required")
	}
	if in.Filename == "" {
		return nil, AddCardAttachmentOutput{}, fmt.Errorf("filename is required")
	}
	data, err := base64.StdEncoding.DecodeString(in.DataBase64)
	if err != nil {
		return nil, AddCardAttachmentOutput{}, fmt.Errorf("data_base64 is not valid base64: %w", err)
	}
	if len(data) == 0 {
		return nil, AddCardAttachmentOutput{}, fmt.Errorf("data_base64 is required")
	}
	if _, err := h.api.AddCardAttachment(ctx, in.CardID, in.Filename, data, in.ContentType); err != nil {
		return nil, AddCardAttachmentOutput{}, err
	}
	return nil, AddCardAttachmentOutput{
		CardID:            in.CardID,
		Filename:          in.Filename,
		MarkdownReference: fmt.Sprintf("![](@media/%s)", in.Filename),
	}, nil
}

func (h *handlers) deleteCardAttachment(ctx context.Context, _ *mcp.CallToolRequest, in DeleteCardAttachmentInput) (*mcp.CallToolResult, DeleteCardAttachmentOutput, error) {
	if in.CardID == "" {
		return nil, DeleteCardAttachmentOutput{}, fmt.Errorf("card_id is required")
	}
	if in.Filename == "" {
		return nil, DeleteCardAttachmentOutput{}, fmt.Errorf("filename is required")
	}
	if err := h.api.DeleteCardAttachment(ctx, in.CardID, in.Filename); err != nil {
		return nil, DeleteCardAttachmentOutput{}, err
	}
	return nil, DeleteCardAttachmentOutput{Deleted: true, CardID: in.CardID, Filename: in.Filename}, nil
}

func (h *handlers) createDeck(ctx context.Context, _ *mcp.CallToolRequest, in CreateDeckInput) (*mcp.CallToolResult, DeckOutput, error) {
	if in.Name == "" {
		return nil, DeckOutput{}, fmt.Errorf("name is required")
	}
	deck, err := h.api.CreateDeck(ctx, mochi.CreateDeckParams{
		Name:     in.Name,
		ParentID: in.ParentID,
	})
	if err != nil {
		return nil, DeckOutput{}, err
	}
	return nil, DeckOutput{Deck: deck}, nil
}
