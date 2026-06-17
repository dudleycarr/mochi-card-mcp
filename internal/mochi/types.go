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

// UpdateDeckParams holds the mutable fields of a deck. Empty strings and a nil
// Sort are omitted from the request, leaving those fields unchanged.
type UpdateDeckParams struct {
	Name     string
	ParentID string
	Sort     *int
}

// TemplateField describes a single field within a template's fields map. Mochi
// keys the map by field ID; ID is repeated inside the value.
type TemplateField struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Type    string         `json:"type,omitempty"`
	Pos     string         `json:"pos,omitempty"`
	Content string         `json:"content,omitempty"`
	Options map[string]any `json:"options,omitempty"`
}

// Template represents a Mochi card template. Content is Markdown with field
// placeholders of the form "<< Field name >>".
type Template struct {
	ID      string                   `json:"id"`
	Name    string                   `json:"name"`
	Content string                   `json:"content"`
	Pos     string                   `json:"pos,omitempty"`
	Fields  map[string]TemplateField `json:"fields,omitempty"`
	Style   map[string]any           `json:"style,omitempty"`
	Options map[string]any           `json:"options,omitempty"`
}

// TemplatesResult is a single page of templates. Bookmark, when non-empty, can
// be passed to the next request to fetch the following page.
type TemplatesResult struct {
	Docs     []Template `json:"docs"`
	Bookmark string     `json:"bookmark,omitempty"`
}

// CreateTemplateParams holds the fields used to create a template.
type CreateTemplateParams struct {
	Name    string
	Content string
	Fields  map[string]TemplateField
}

// DueCardsParams holds the optional filters for listing due cards. DeckID, when
// set, restricts the result to a single deck. Date, when set, returns the cards
// due on that date (a timestamp); when empty, Mochi uses today's date.
type DueCardsParams struct {
	DeckID string
	Date   string
}
