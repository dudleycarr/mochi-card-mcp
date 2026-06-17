package mochi

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// cardBody is the JSON request body for creating and updating cards. Mochi uses
// hyphenated keys such as "deck-id".
type cardBody struct {
	Content string `json:"content,omitempty"`
	DeckID  string `json:"deck-id,omitempty"`
}

// ListCards returns a single page of cards, honoring the optional deck filter,
// pagination bookmark, and page size.
func (c *Client) ListCards(ctx context.Context, params ListCardsParams) (CardsResult, error) {
	q := url.Values{}
	if params.DeckID != "" {
		q.Set("deck-id", params.DeckID)
	}
	if params.Bookmark != "" {
		q.Set("bookmark", params.Bookmark)
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}

	var result CardsResult
	if err := c.do(ctx, "GET", "/cards", q, nil, &result); err != nil {
		return CardsResult{}, err
	}
	return result, nil
}

// GetCard returns a single card by ID.
func (c *Client) GetCard(ctx context.Context, id string) (Card, error) {
	var card Card
	if err := c.do(ctx, "GET", "/cards/"+url.PathEscape(id), nil, nil, &card); err != nil {
		return Card{}, err
	}
	return card, nil
}

// CreateCard creates a new card and returns the created card.
func (c *Client) CreateCard(ctx context.Context, params CreateCardParams) (Card, error) {
	var card Card
	body := cardBody{Content: params.Content, DeckID: params.DeckID}
	if err := c.do(ctx, "POST", "/cards", nil, body, &card); err != nil {
		return Card{}, err
	}
	return card, nil
}

// UpdateCard updates an existing card's content and returns the updated card.
func (c *Client) UpdateCard(ctx context.Context, id string, params UpdateCardParams) (Card, error) {
	var card Card
	body := cardBody{Content: params.Content}
	if err := c.do(ctx, "POST", "/cards/"+url.PathEscape(id), nil, body, &card); err != nil {
		return Card{}, err
	}
	return card, nil
}

// DeleteCard permanently deletes a card by ID.
func (c *Client) DeleteCard(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", "/cards/"+url.PathEscape(id), nil, nil, nil)
}

// SearchCards returns cards whose content matches the query (case-insensitive
// substring match). Mochi has no server-side search endpoint, so this filters a
// single page of cards client-side. The returned Bookmark, when non-empty, can
// be passed back to continue scanning further pages.
func (c *Client) SearchCards(ctx context.Context, query, bookmark string) (CardsResult, error) {
	page, err := c.ListCards(ctx, ListCardsParams{Bookmark: bookmark, Limit: 100})
	if err != nil {
		return CardsResult{}, err
	}

	needle := strings.ToLower(query)
	matches := make([]Card, 0, len(page.Docs))
	for _, card := range page.Docs {
		if needle == "" || strings.Contains(strings.ToLower(card.Content), needle) {
			matches = append(matches, card)
		}
	}
	return CardsResult{Docs: matches, Bookmark: page.Bookmark}, nil
}
