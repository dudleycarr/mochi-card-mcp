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

// CreateDeck creates a new deck and returns it.
func (c *Client) CreateDeck(ctx context.Context, params CreateDeckParams) (Deck, error) {
	var deck Deck
	body := deckBody{Name: params.Name, ParentID: params.ParentID}
	if err := c.do(ctx, "POST", "/decks", nil, body, &deck); err != nil {
		return Deck{}, err
	}
	return deck, nil
}
