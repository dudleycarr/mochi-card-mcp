package mochi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddCardAttachment(t *testing.T) {
	var gotMethod, gotPath, gotPartName, gotFilename, gotPartContentType, gotData string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("content-type = %q, want multipart/form-data", r.Header.Get("Content-Type"))
		}
		mr, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("multipart: %v", err)
		}
		p, err := mr.NextPart()
		if err != nil {
			t.Fatalf("next part: %v", err)
		}
		b, _ := io.ReadAll(p)
		gotPartName, gotFilename, gotPartContentType, gotData = p.FormName(), p.FileName(), p.Header.Get("Content-Type"), string(b)
		io.WriteString(w, `{"id":"c1"}`)
	}))
	defer srv.Close()

	card, err := newTestClient(srv).AddCardAttachment(context.Background(), "c1", "note.txt", []byte("hello"), "text/plain")
	if err != nil {
		t.Fatalf("AddCardAttachment: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/cards/c1/attachments/note.txt" {
		t.Errorf("got %s %s, want POST /cards/c1/attachments/note.txt", gotMethod, gotPath)
	}
	if gotPartName != "file" {
		t.Errorf("form field = %q, want file", gotPartName)
	}
	if gotFilename != "note.txt" {
		t.Errorf("filename = %q, want note.txt", gotFilename)
	}
	if gotPartContentType != "text/plain" {
		t.Errorf("part content-type = %q, want text/plain", gotPartContentType)
	}
	if gotData != "hello" {
		t.Errorf("uploaded data = %q, want hello", gotData)
	}
	if card.ID != "c1" {
		t.Errorf("card.ID = %q, want c1", card.ID)
	}
}

func TestAddCardAttachmentWithoutContentType(t *testing.T) {
	var gotPartContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mr, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("multipart: %v", err)
		}
		p, err := mr.NextPart()
		if err != nil {
			t.Fatalf("next part: %v", err)
		}
		io.ReadAll(p)
		gotPartContentType = p.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if _, err := newTestClient(srv).AddCardAttachment(context.Background(), "c1", "f.bin", []byte("x"), ""); err != nil {
		t.Fatalf("AddCardAttachment: %v", err)
	}
	if gotPartContentType != "" {
		t.Errorf("part content-type = %q, want empty", gotPartContentType)
	}
}

func TestDeleteCardAttachment(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := newTestClient(srv).DeleteCardAttachment(context.Background(), "c1", "old.png"); err != nil {
		t.Fatalf("DeleteCardAttachment: %v", err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/cards/c1/attachments/old.png" {
		t.Errorf("got %s %s, want DELETE /cards/c1/attachments/old.png", gotMethod, gotPath)
	}
}
