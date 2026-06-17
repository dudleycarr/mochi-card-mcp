package mochi

// Card represents a Mochi flashcard. Mochi stores both sides of a card in a
// single Markdown Content field, with the two sides separated by a line
// containing only "---".
type Card struct {
	ID         string   `json:"id"`
	Content    string   `json:"content"`
	DeckID     string   `json:"deck-id,omitempty"`
	TemplateID string   `json:"template-id,omitempty"`
	Pos        string   `json:"pos,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	ManualTags []string `json:"manual-tags,omitempty"`
	// CreatedAt and UpdatedAt are passed through from the API as-is; Mochi
	// returns them as objects, so they are left untyped.
	CreatedAt any `json:"created-at,omitempty"`
	UpdatedAt any `json:"updated-at,omitempty"`
}

// Deck represents a Mochi deck, a named collection of cards.
type Deck struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent-id,omitempty"`
	Sort     int    `json:"sort,omitempty"`
}

// CardsResult is a single page of cards returned by list and search
// operations. Bookmark, when non-empty, can be passed to the next request to
// fetch the following page.
type CardsResult struct {
	Docs     []Card `json:"docs"`
	Bookmark string `json:"bookmark,omitempty"`
}

// DecksResult is a single page of decks. Bookmark, when non-empty, can be
// passed to the next request to fetch the following page.
type DecksResult struct {
	Docs     []Deck `json:"docs"`
	Bookmark string `json:"bookmark,omitempty"`
}

// ListCardsParams holds the optional filters for listing cards.
type ListCardsParams struct {
	DeckID   string
	Bookmark string
	Limit    int
}

// CreateCardParams holds the fields used to create a card. Content is the full
// Markdown body (both sides separated by "---"); DeckID is optional.
type CreateCardParams struct {
	Content string
	DeckID  string
}

// UpdateCardParams holds the mutable fields of a card. Only Content is
// updatable through this client.
type UpdateCardParams struct {
	Content string
}

// CreateDeckParams holds the fields used to create a deck.
type CreateDeckParams struct {
	Name     string
	ParentID string
}

// DueCardsParams holds the optional filters for listing due cards. DeckID, when
// set, restricts the result to a single deck. Date, when set, returns the cards
// due on that date (a timestamp); when empty, Mochi uses today's date.
type DueCardsParams struct {
	DeckID string
	Date   string
}
