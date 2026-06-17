package server

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/dudleycarr/mochi-card-mcp/internal/mochi"
)

// fakeAPI is a configurable in-memory implementation of cardAPI.
type fakeAPI struct {
	cards map[string]mochi.Card
	decks []mochi.Deck

	listCardsParams  mochi.ListCardsParams
	dueCardsParams   mochi.DueCardsParams
	createCardParams mochi.CreateCardParams
	updateCardParams mochi.UpdateCardParams
	updateCardID     string
	deletedID        string
	searchQuery      string
	createDeckParams mochi.CreateDeckParams
	updateDeckParams mochi.UpdateDeckParams
	updateDeckID     string
	deletedDeckID    string

	err error
}

func (f *fakeAPI) ListCards(_ context.Context, params mochi.ListCardsParams) (mochi.CardsResult, error) {
	f.listCardsParams = params
	if f.err != nil {
		return mochi.CardsResult{}, f.err
	}
	var docs []mochi.Card
	for _, c := range f.cards {
		docs = append(docs, c)
	}
	return mochi.CardsResult{Docs: docs, Bookmark: "bm"}, nil
}

func (f *fakeAPI) ListDueCards(_ context.Context, params mochi.DueCardsParams) ([]mochi.Card, error) {
	f.dueCardsParams = params
	if f.err != nil {
		return nil, f.err
	}
	return []mochi.Card{{ID: "due1", Content: "due card"}}, nil
}

func (f *fakeAPI) GetCard(_ context.Context, id string) (mochi.Card, error) {
	if f.err != nil {
		return mochi.Card{}, f.err
	}
	c, ok := f.cards[id]
	if !ok {
		return mochi.Card{}, errors.New("not found")
	}
	return c, nil
}

func (f *fakeAPI) CreateCard(_ context.Context, params mochi.CreateCardParams) (mochi.Card, error) {
	f.createCardParams = params
	if f.err != nil {
		return mochi.Card{}, f.err
	}
	return mochi.Card{ID: "new-card", Content: params.Content, DeckID: params.DeckID}, nil
}

func (f *fakeAPI) UpdateCard(_ context.Context, id string, params mochi.UpdateCardParams) (mochi.Card, error) {
	f.updateCardID = id
	f.updateCardParams = params
	if f.err != nil {
		return mochi.Card{}, f.err
	}
	return mochi.Card{ID: id, Content: params.Content}, nil
}

func (f *fakeAPI) DeleteCard(_ context.Context, id string) error {
	f.deletedID = id
	return f.err
}

func (f *fakeAPI) SearchCards(_ context.Context, query, bookmark string) (mochi.CardsResult, error) {
	f.searchQuery = query
	if f.err != nil {
		return mochi.CardsResult{}, f.err
	}
	return mochi.CardsResult{Docs: []mochi.Card{{ID: "match", Content: query}}, Bookmark: bookmark}, nil
}

func (f *fakeAPI) ListDecks(_ context.Context, _ string) (mochi.DecksResult, error) {
	if f.err != nil {
		return mochi.DecksResult{}, f.err
	}
	return mochi.DecksResult{Docs: f.decks}, nil
}

func (f *fakeAPI) GetDeck(_ context.Context, id string) (mochi.Deck, error) {
	if f.err != nil {
		return mochi.Deck{}, f.err
	}
	for _, d := range f.decks {
		if d.ID == id {
			return d, nil
		}
	}
	return mochi.Deck{}, errors.New("not found")
}

func (f *fakeAPI) UpdateDeck(_ context.Context, id string, params mochi.UpdateDeckParams) (mochi.Deck, error) {
	f.updateDeckID = id
	f.updateDeckParams = params
	if f.err != nil {
		return mochi.Deck{}, f.err
	}
	return mochi.Deck{ID: id, Name: params.Name, ParentID: params.ParentID}, nil
}

func (f *fakeAPI) DeleteDeck(_ context.Context, id string) error {
	f.deletedDeckID = id
	return f.err
}

func (f *fakeAPI) CreateDeck(_ context.Context, params mochi.CreateDeckParams) (mochi.Deck, error) {
	f.createDeckParams = params
	if f.err != nil {
		return mochi.Deck{}, f.err
	}
	return mochi.Deck{ID: "new-deck", Name: params.Name, ParentID: params.ParentID}, nil
}

func TestCreateCardJoinsSides(t *testing.T) {
	f := &fakeAPI{}
	h := &handlers{api: f}

	_, out, err := h.createCard(context.Background(), nil, CreateCardInput{Name: "Front", Content: "Back", DeckID: "d1"})
	if err != nil {
		t.Fatalf("createCard: %v", err)
	}
	if f.createCardParams.Content != "Front\n---\nBack" {
		t.Errorf("content = %q, want joined sides", f.createCardParams.Content)
	}
	if f.createCardParams.DeckID != "d1" {
		t.Errorf("deck = %q, want d1", f.createCardParams.DeckID)
	}
	if out.Card.ID != "new-card" {
		t.Errorf("card.ID = %q", out.Card.ID)
	}
}

func TestListDueCards(t *testing.T) {
	f := &fakeAPI{}
	h := &handlers{api: f}
	_, out, err := h.listDueCards(context.Background(), nil, ListDueCardsInput{DeckID: "d1", Date: "2026-06-17"})
	if err != nil {
		t.Fatalf("listDueCards: %v", err)
	}
	if f.dueCardsParams.DeckID != "d1" || f.dueCardsParams.Date != "2026-06-17" {
		t.Errorf("unexpected params: %+v", f.dueCardsParams)
	}
	if len(out.Cards) != 1 || out.Cards[0].ID != "due1" {
		t.Errorf("unexpected cards: %+v", out.Cards)
	}
}

func TestCreateCardRequiresName(t *testing.T) {
	h := &handlers{api: &fakeAPI{}}
	if _, _, err := h.createCard(context.Background(), nil, CreateCardInput{Content: "Back"}); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestUpdateCardMergesOneSide(t *testing.T) {
	f := &fakeAPI{cards: map[string]mochi.Card{
		"c1": {ID: "c1", Content: "OldFront\n---\nOldBack"},
	}}
	h := &handlers{api: f}

	// Update only the back; the front should be preserved.
	_, _, err := h.updateCard(context.Background(), nil, UpdateCardInput{CardID: "c1", Content: "NewBack"})
	if err != nil {
		t.Fatalf("updateCard: %v", err)
	}
	if want := "OldFront\n---\nNewBack"; f.updateCardParams.Content != want {
		t.Errorf("content = %q, want %q", f.updateCardParams.Content, want)
	}
	if f.updateCardID != "c1" {
		t.Errorf("updated id = %q, want c1", f.updateCardID)
	}
}

func TestUpdateCardRequiresAField(t *testing.T) {
	h := &handlers{api: &fakeAPI{}}
	if _, _, err := h.updateCard(context.Background(), nil, UpdateCardInput{CardID: "c1"}); err == nil {
		t.Fatal("expected error when no fields provided")
	}
}

func TestDeleteCard(t *testing.T) {
	f := &fakeAPI{}
	h := &handlers{api: f}
	_, out, err := h.deleteCard(context.Background(), nil, DeleteCardInput{CardID: "c9"})
	if err != nil {
		t.Fatalf("deleteCard: %v", err)
	}
	if f.deletedID != "c9" {
		t.Errorf("deleted id = %q, want c9", f.deletedID)
	}
	if !out.Deleted || out.CardID != "c9" {
		t.Errorf("unexpected output: %+v", out)
	}
}

func TestSearchCardsPassesQuery(t *testing.T) {
	f := &fakeAPI{}
	h := &handlers{api: f}
	_, out, err := h.searchCards(context.Background(), nil, SearchCardsInput{Query: "fox", Bookmark: "bm"})
	if err != nil {
		t.Fatalf("searchCards: %v", err)
	}
	if f.searchQuery != "fox" {
		t.Errorf("query = %q, want fox", f.searchQuery)
	}
	if len(out.Cards) != 1 || out.Cards[0].ID != "match" {
		t.Errorf("unexpected cards: %+v", out.Cards)
	}
}

func TestCreateDeck(t *testing.T) {
	f := &fakeAPI{}
	h := &handlers{api: f}
	_, out, err := h.createDeck(context.Background(), nil, CreateDeckInput{Name: "Spanish", ParentID: "p1"})
	if err != nil {
		t.Fatalf("createDeck: %v", err)
	}
	if f.createDeckParams.Name != "Spanish" || f.createDeckParams.ParentID != "p1" {
		t.Errorf("unexpected params: %+v", f.createDeckParams)
	}
	if out.Deck.ID != "new-deck" {
		t.Errorf("deck.ID = %q", out.Deck.ID)
	}
}

func TestGetDeck(t *testing.T) {
	f := &fakeAPI{decks: []mochi.Deck{{ID: "d1", Name: "Spanish"}}}
	h := &handlers{api: f}
	_, out, err := h.getDeck(context.Background(), nil, GetDeckInput{DeckID: "d1"})
	if err != nil {
		t.Fatalf("getDeck: %v", err)
	}
	if out.Deck.ID != "d1" || out.Deck.Name != "Spanish" {
		t.Errorf("unexpected deck: %+v", out.Deck)
	}
}

func TestGetDeckRequiresID(t *testing.T) {
	h := &handlers{api: &fakeAPI{}}
	if _, _, err := h.getDeck(context.Background(), nil, GetDeckInput{}); err == nil {
		t.Fatal("expected error for missing deck_id")
	}
}

func TestUpdateDeck(t *testing.T) {
	f := &fakeAPI{}
	h := &handlers{api: f}
	sort := 3
	_, out, err := h.updateDeck(context.Background(), nil, UpdateDeckInput{DeckID: "d1", Name: "Español", Sort: &sort})
	if err != nil {
		t.Fatalf("updateDeck: %v", err)
	}
	if f.updateDeckID != "d1" {
		t.Errorf("updated id = %q, want d1", f.updateDeckID)
	}
	if f.updateDeckParams.Name != "Español" || f.updateDeckParams.Sort == nil || *f.updateDeckParams.Sort != 3 {
		t.Errorf("unexpected params: %+v", f.updateDeckParams)
	}
	if out.Deck.ID != "d1" {
		t.Errorf("deck.ID = %q", out.Deck.ID)
	}
}

func TestUpdateDeckRequiresAField(t *testing.T) {
	h := &handlers{api: &fakeAPI{}}
	if _, _, err := h.updateDeck(context.Background(), nil, UpdateDeckInput{DeckID: "d1"}); err == nil {
		t.Fatal("expected error when no fields provided")
	}
}

func TestDeleteDeck(t *testing.T) {
	f := &fakeAPI{}
	h := &handlers{api: f}
	_, out, err := h.deleteDeck(context.Background(), nil, DeleteDeckInput{DeckID: "d9"})
	if err != nil {
		t.Fatalf("deleteDeck: %v", err)
	}
	if f.deletedDeckID != "d9" {
		t.Errorf("deleted id = %q, want d9", f.deletedDeckID)
	}
	if !out.Deleted || out.DeckID != "d9" {
		t.Errorf("unexpected output: %+v", out)
	}
}

func TestHandlerPropagatesError(t *testing.T) {
	h := &handlers{api: &fakeAPI{err: errors.New("boom")}}
	if _, _, err := h.listDecks(context.Background(), nil, ListDecksInput{}); err == nil {
		t.Fatal("expected error to propagate")
	}
}

// TestServerToolsRegistered exercises the full MCP stack over an in-memory
// transport, confirming the tools are registered and callable end to end.
func TestServerToolsRegistered(t *testing.T) {
	f := &fakeAPI{
		decks: []mochi.Deck{{ID: "d1", Name: "Spanish"}},
	}
	srv := newServer(f, "test")

	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := srv.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer clientSession.Close()

	tools, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	want := map[string]bool{
		"mochi_list_cards": true, "mochi_list_due_cards": true, "mochi_get_card": true,
		"mochi_create_card": true, "mochi_update_card": true, "mochi_delete_card": true,
		"mochi_search_cards": true, "mochi_list_decks": true, "mochi_get_deck": true,
		"mochi_create_deck": true, "mochi_update_deck": true, "mochi_delete_deck": true,
	}
	got := map[string]bool{}
	for _, tool := range tools.Tools {
		got[tool.Name] = true
	}
	for name := range want {
		if !got[name] {
			t.Errorf("tool %q not registered", name)
		}
	}
	if len(tools.Tools) != len(want) {
		t.Errorf("got %d tools, want %d", len(tools.Tools), len(want))
	}

	// Call one tool end to end.
	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "mochi_list_decks",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("call mochi_list_decks: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %+v", res.Content)
	}
	out, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content type = %T", res.StructuredContent)
	}
	decks, ok := out["decks"].([]any)
	if !ok || len(decks) != 1 {
		t.Fatalf("unexpected decks in output: %+v", out)
	}
}
