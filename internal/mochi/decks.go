package mochi

import (
	"context"
	"net/url"
)

// deckBody is the JSON request body for creating decks.
type deckBody struct {
	Name     string `json:"name"`
	ParentID string `json:"parent-id,omitempty"`
}

// updateDeckBody is the JSON request body for updating decks. Unset fields are
// omitted so they are left unchanged.
type updateDeckBody struct {
	Name     string `json:"name,omitempty"`
	ParentID string `json:"parent-id,omitempty"`
	Sort     *int   `json:"sort,omitempty"`
}

// ListDecks returns a single page of decks. Pass the returned Bookmark to fetch
// the next page.
func (c *Client) ListDecks(ctx context.Context, bookmark string) (DecksResult, error) {
	q := url.Values{}
	if bookmark != "" {
		q.Set("bookmark", bookmark)
	}

	var result DecksResult
	if err := c.do(ctx, "GET", "/decks", q, nil, &result); err != nil {
		return DecksResult{}, err
	}
	return result, nil
}

// GetDeck returns a single deck by ID.
func (c *Client) GetDeck(ctx context.Context, id string) (Deck, error) {
	var deck Deck
	if err := c.do(ctx, "GET", "/decks/"+url.PathEscape(id), nil, nil, &deck); err != nil {
		return Deck{}, err
	}
	return deck, nil
}

// CreateDeck creates a new deck and returns it.
func (c *Client) CreateDeck(ctx context.Context, params CreateDeckParams) (Deck, error) {
	var deck Deck
	body := deckBody{Name: params.Name, ParentID: params.ParentID}
	if err := c.do(ctx, "POST", "/decks", nil, body, &deck); err != nil {
		return Deck{}, err
	}
	return deck, nil
}

// UpdateDeck updates an existing deck and returns it.
func (c *Client) UpdateDeck(ctx context.Context, id string, params UpdateDeckParams) (Deck, error) {
	var deck Deck
	body := updateDeckBody{Name: params.Name, ParentID: params.ParentID, Sort: params.Sort}
	if err := c.do(ctx, "POST", "/decks/"+url.PathEscape(id), nil, body, &deck); err != nil {
		return Deck{}, err
	}
	return deck, nil
}

// DeleteDeck permanently deletes a deck by ID.
func (c *Client) DeleteDeck(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", "/decks/"+url.PathEscape(id), nil, nil, nil)
}
