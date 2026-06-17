package mochi

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
)

// attachmentPath builds the path for a card attachment, escaping both segments.
func attachmentPath(cardID, filename string) string {
	return "/cards/" + url.PathEscape(cardID) + "/attachments/" + url.PathEscape(filename)
}

// AddCardAttachment uploads an attachment to a card. Mochi expects the file as
// multipart/form-data under a field named "file". contentType is optional; when
// empty the part is sent without an explicit Content-Type. The returned Card is
// best-effort: Mochi does not document the response body, so it may be empty.
//
// Once uploaded, an attachment is referenced from card Markdown as
// "![](@media/<filename>)".
func (c *Client) AddCardAttachment(ctx context.Context, cardID, filename string, data []byte, contentType string) (Card, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename=%q`, filename))
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	part, err := mw.CreatePart(header)
	if err != nil {
		return Card{}, fmt.Errorf("mochi: building multipart form: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return Card{}, fmt.Errorf("mochi: writing attachment data: %w", err)
	}
	if err := mw.Close(); err != nil {
		return Card{}, fmt.Errorf("mochi: finalizing multipart form: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+attachmentPath(cardID, filename), &buf)
	if err != nil {
		return Card{}, fmt.Errorf("mochi: building request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	var card Card
	if err := c.doRequest(req, &card); err != nil {
		return Card{}, err
	}
	return card, nil
}

// DeleteCardAttachment removes an attachment from a card by filename.
func (c *Client) DeleteCardAttachment(ctx context.Context, cardID, filename string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+attachmentPath(cardID, filename), nil)
	if err != nil {
		return fmt.Errorf("mochi: building request: %w", err)
	}
	return c.doRequest(req, nil)
}
