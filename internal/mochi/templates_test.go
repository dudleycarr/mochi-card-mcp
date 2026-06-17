package mochi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListTemplates(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		io.WriteString(w, `{"docs":[{"id":"t1","name":"Basic","content":"<< Front >>","fields":{"front":{"id":"front","name":"Front","type":"text"}}}],"bookmark":"next"}`)
	}))
	defer srv.Close()

	res, err := newTestClient(srv).ListTemplates(context.Background(), "bm")
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}
	if gotPath != "/templates" {
		t.Errorf("path = %q, want /templates", gotPath)
	}
	if gotQuery != "bookmark=bm" {
		t.Errorf("query = %q, want bookmark=bm", gotQuery)
	}
	if len(res.Docs) != 1 || res.Docs[0].ID != "t1" {
		t.Fatalf("unexpected docs: %+v", res.Docs)
	}
	if f := res.Docs[0].Fields["front"]; f.Type != "text" || f.Name != "Front" {
		t.Errorf("field not parsed: %+v", res.Docs[0].Fields)
	}
	if res.Bookmark != "next" {
		t.Errorf("bookmark = %q, want next", res.Bookmark)
	}
}

func TestGetTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/templates/t1" {
			t.Errorf("got %s %s, want GET /templates/t1", r.Method, r.URL.Path)
		}
		io.WriteString(w, `{"id":"t1","name":"Basic","content":"<< Front >>"}`)
	}))
	defer srv.Close()

	tmpl, err := newTestClient(srv).GetTemplate(context.Background(), "t1")
	if err != nil {
		t.Fatalf("GetTemplate: %v", err)
	}
	if tmpl.ID != "t1" || tmpl.Name != "Basic" {
		t.Errorf("unexpected template: %+v", tmpl)
	}
}

func TestCreateTemplate(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/templates" {
			t.Errorf("got %s %s, want POST /templates", r.Method, r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		io.WriteString(w, `{"id":"new","name":"Vocab","content":"<< Word >>"}`)
	}))
	defer srv.Close()

	tmpl, err := newTestClient(srv).CreateTemplate(context.Background(), CreateTemplateParams{
		Name:    "Vocab",
		Content: "<< Word >>",
		Fields: map[string]TemplateField{
			"word": {ID: "word", Name: "Word", Type: "text"},
		},
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	if gotBody["name"] != "Vocab" || gotBody["content"] != "<< Word >>" {
		t.Errorf("unexpected body: %+v", gotBody)
	}
	fields, ok := gotBody["fields"].(map[string]any)
	if !ok {
		t.Fatalf("fields missing or wrong type: %+v", gotBody["fields"])
	}
	word, ok := fields["word"].(map[string]any)
	if !ok || word["type"] != "text" {
		t.Errorf("field not encoded: %+v", fields)
	}
	if tmpl.ID != "new" {
		t.Errorf("template.ID = %q, want new", tmpl.ID)
	}
}
