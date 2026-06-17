package mochi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient returns a Client pointed at the given test server.
func newTestClient(srv *httptest.Server) *Client {
	return NewClient("test-key", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
}

func TestClientSetsBasicAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		io.WriteString(w, `{"docs":[]}`)
	}))
	defer srv.Close()

	if _, err := newTestClient(srv).ListDecks(context.Background(), ""); err != nil {
		t.Fatalf("ListDecks: %v", err)
	}

	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-key:"))
	if gotAuth != want {
		t.Errorf("Authorization = %q, want %q", gotAuth, want)
	}
}

func TestListCards(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		io.WriteString(w, `{"docs":[{"id":"c1","content":"Front\n---\nBack","deck-id":"d1"}],"bookmark":"next"}`)
	}))
	defer srv.Close()

	res, err := newTestClient(srv).ListCards(context.Background(), ListCardsParams{DeckID: "d1", Bookmark: "bm", Limit: 25})
	if err != nil {
		t.Fatalf("ListCards: %v", err)
	}
	if gotPath != "/cards" {
		t.Errorf("path = %q, want /cards", gotPath)
	}
	for _, want := range []string{"deck-id=d1", "bookmark=bm", "limit=25"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
	if len(res.Docs) != 1 || res.Docs[0].ID != "c1" {
		t.Fatalf("unexpected docs: %+v", res.Docs)
	}
	if res.Bookmark != "next" {
		t.Errorf("bookmark = %q, want next", res.Bookmark)
	}
}

func TestGetCard(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cards/c1" {
			t.Errorf("path = %q, want /cards/c1", r.URL.Path)
		}
		io.WriteString(w, `{"id":"c1","content":"hi"}`)
	}))
	defer srv.Close()

	card, err := newTestClient(srv).GetCard(context.Background(), "c1")
	if err != nil {
		t.Fatalf("GetCard: %v", err)
	}
	if card.ID != "c1" || card.Content != "hi" {
		t.Errorf("unexpected card: %+v", card)
	}
}

func TestCreateCard(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		io.WriteString(w, `{"id":"new","content":"Front\n---\nBack","deck-id":"d1"}`)
	}))
	defer srv.Close()

	card, err := newTestClient(srv).CreateCard(context.Background(), CreateCardParams{Content: "Front\n---\nBack", DeckID: "d1"})
	if err != nil {
		t.Fatalf("CreateCard: %v", err)
	}
	if gotBody["content"] != "Front\n---\nBack" {
		t.Errorf("body content = %v", gotBody["content"])
	}
	if gotBody["deck-id"] != "d1" {
		t.Errorf("body deck-id = %v, want d1", gotBody["deck-id"])
	}
	if card.ID != "new" {
		t.Errorf("card.ID = %q, want new", card.ID)
	}
}

func TestUpdateCard(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cards/c1" {
			t.Errorf("path = %q, want /cards/c1", r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		io.WriteString(w, `{"id":"c1","content":"new content"}`)
	}))
	defer srv.Close()

	card, err := newTestClient(srv).UpdateCard(context.Background(), "c1", UpdateCardParams{Content: "new content"})
	if err != nil {
		t.Fatalf("UpdateCard: %v", err)
	}
	if gotBody["content"] != "new content" {
		t.Errorf("body content = %v", gotBody["content"])
	}
	if card.Content != "new content" {
		t.Errorf("card.Content = %q", card.Content)
	}
}

func TestDeleteCard(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := newTestClient(srv).DeleteCard(context.Background(), "c1"); err != nil {
		t.Fatalf("DeleteCard: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/cards/c1" {
		t.Errorf("got %s %s, want DELETE /cards/c1", gotMethod, gotPath)
	}
}

func TestSearchCardsFiltersClientSide(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"docs":[
			{"id":"c1","content":"the quick brown fox"},
			{"id":"c2","content":"lazy dog"},
			{"id":"c3","content":"QUICK silver"}
		],"bookmark":"next"}`)
	}))
	defer srv.Close()

	res, err := newTestClient(srv).SearchCards(context.Background(), "quick", "")
	if err != nil {
		t.Fatalf("SearchCards: %v", err)
	}
	if len(res.Docs) != 2 {
		t.Fatalf("got %d matches, want 2: %+v", len(res.Docs), res.Docs)
	}
	if res.Docs[0].ID != "c1" || res.Docs[1].ID != "c3" {
		t.Errorf("unexpected matches: %+v", res.Docs)
	}
	if res.Bookmark != "next" {
		t.Errorf("bookmark = %q, want next", res.Bookmark)
	}
}

func TestListDecksAndCreateDeck(t *testing.T) {
	var createBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/decks":
			io.WriteString(w, `{"docs":[{"id":"d1","name":"Spanish"}]}`)
		case r.Method == http.MethodPost && r.URL.Path == "/decks":
			json.NewDecoder(r.Body).Decode(&createBody)
			io.WriteString(w, `{"id":"d2","name":"French","parent-id":"d1"}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv)
	decks, err := c.ListDecks(context.Background(), "")
	if err != nil {
		t.Fatalf("ListDecks: %v", err)
	}
	if len(decks.Docs) != 1 || decks.Docs[0].Name != "Spanish" {
		t.Errorf("unexpected decks: %+v", decks.Docs)
	}

	deck, err := c.CreateDeck(context.Background(), CreateDeckParams{Name: "French", ParentID: "d1"})
	if err != nil {
		t.Fatalf("CreateDeck: %v", err)
	}
	if createBody["name"] != "French" || createBody["parent-id"] != "d1" {
		t.Errorf("unexpected create body: %+v", createBody)
	}
	if deck.ID != "d2" {
		t.Errorf("deck.ID = %q, want d2", deck.ID)
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "invalid api key")
	}))
	defer srv.Close()

	_, err := newTestClient(srv).GetCard(context.Background(), "c1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
	if apiErr.Body != "invalid api key" {
		t.Errorf("Body = %q", apiErr.Body)
	}
}
