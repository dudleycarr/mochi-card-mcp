// Package server wires the Mochi Cards API to MCP tools exposed over the
// Model Context Protocol.
package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/dudleycarr/mochi-card-mcp/internal/mochi"
)

// cardAPI is the subset of the Mochi client used by the tool handlers. It is an
// interface so the handlers can be tested without a live API.
type cardAPI interface {
	ListCards(ctx context.Context, params mochi.ListCardsParams) (mochi.CardsResult, error)
	GetCard(ctx context.Context, id string) (mochi.Card, error)
	CreateCard(ctx context.Context, params mochi.CreateCardParams) (mochi.Card, error)
	UpdateCard(ctx context.Context, id string, params mochi.UpdateCardParams) (mochi.Card, error)
	DeleteCard(ctx context.Context, id string) error
	SearchCards(ctx context.Context, query, bookmark string) (mochi.CardsResult, error)
	ListDecks(ctx context.Context, bookmark string) (mochi.DecksResult, error)
	CreateDeck(ctx context.Context, params mochi.CreateDeckParams) (mochi.Deck, error)
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
		Name:        "mochi_create_deck",
		Description: "Create a Mochi deck with the given name and optional parent deck.",
	}, h.createDeck)
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

// CreateDeckInput are the arguments for mochi_create_deck.
type CreateDeckInput struct {
	Name     string `json:"name" jsonschema:"the name of the deck"`
	ParentID string `json:"parent_id,omitempty" jsonschema:"optional parent deck ID"`
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
