package mochi

import (
	"context"
	"net/url"
)

// templateBody is the JSON request body for creating templates.
type templateBody struct {
	Name    string                   `json:"name"`
	Content string                   `json:"content"`
	Fields  map[string]TemplateField `json:"fields,omitempty"`
}

// ListTemplates returns a single page of templates. Pass the returned Bookmark
// to fetch the next page.
func (c *Client) ListTemplates(ctx context.Context, bookmark string) (TemplatesResult, error) {
	q := url.Values{}
	if bookmark != "" {
		q.Set("bookmark", bookmark)
	}

	var result TemplatesResult
	if err := c.do(ctx, "GET", "/templates", q, nil, &result); err != nil {
		return TemplatesResult{}, err
	}
	return result, nil
}

// GetTemplate returns a single template by ID.
func (c *Client) GetTemplate(ctx context.Context, id string) (Template, error) {
	var tmpl Template
	if err := c.do(ctx, "GET", "/templates/"+url.PathEscape(id), nil, nil, &tmpl); err != nil {
		return Template{}, err
	}
	return tmpl, nil
}

// CreateTemplate creates a new template and returns it.
func (c *Client) CreateTemplate(ctx context.Context, params CreateTemplateParams) (Template, error) {
	var tmpl Template
	body := templateBody{Name: params.Name, Content: params.Content, Fields: params.Fields}
	if err := c.do(ctx, "POST", "/templates", nil, body, &tmpl); err != nil {
		return Template{}, err
	}
	return tmpl, nil
}
