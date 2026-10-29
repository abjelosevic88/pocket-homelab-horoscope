// Package pocket talks to sibling Pocket apps: Pocket Contacts (birthdays
// become people), Pocket Notes (a reading kept as a note) and Pocket Health
// (sleep and heart rate next to the Moon). Same bearer-token JSON API everywhere.
package pocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a bearer-token JSON client for one Pocket app.
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

// New creates a client.
func New(baseURL, token string, timeout time.Duration) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), Token: token, HTTP: &http.Client{Timeout: timeout}}
}

// Configured reports whether URL and token are set.
func (c *Client) Configured() bool { return c != nil && c.BaseURL != "" && c.Token != "" }

// ErrNotFound is a 404 from the sibling app.
var ErrNotFound = errors.New("not found")

func (c *Client) do(ctx context.Context, method, path string, q url.Values, body any, v any) error {
	if !c.Configured() {
		return errors.New("not configured")
	}
	u := c.BaseURL + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusUnauthorized {
		return errors.New("token rejected (401)")
	}
	if res.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if res.StatusCode >= 300 {
		var e struct {
			Error string `json:"error"`
		}
		b, _ := io.ReadAll(io.LimitReader(res.Body, 2000))
		_ = json.Unmarshal(b, &e)
		if e.Error != "" {
			return fmt.Errorf("HTTP %d: %s", res.StatusCode, e.Error)
		}
		return fmt.Errorf("HTTP %d", res.StatusCode)
	}
	if v == nil || res.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(v)
}

// Health is the connection test for any Pocket app: /healthz plus a token check.
func (c *Client) Health(ctx context.Context) (string, error) {
	var out struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	if !c.Configured() {
		return "", errors.New("not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/healthz", nil)
	if err != nil {
		return "", err
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", res.StatusCode)
	}
	_ = json.NewDecoder(res.Body).Decode(&out)
	var me map[string]any
	if err := c.do(ctx, http.MethodGet, "/api/v1/auth/me", nil, nil, &me); err != nil {
		return out.Version, err
	}
	return out.Version, nil
}

// ---------------- Pocket Contacts ----------------

// Contact is a Pocket Contacts contact (the fields Horoscope reads).
type Contact struct {
	ID       int64  `json:"id"`
	UID      string `json:"uid"`
	Kind     string `json:"kind"`
	FN       string `json:"fn"`
	Given    string `json:"given"`
	Family   string `json:"family"`
	Nickname string `json:"nickname"`
	Birthday string `json:"birthday"` // YYYY-MM-DD, or --MM-DD without a year
	Color    string `json:"color"`
}

// Contacts lists the person's contacts (up to 5000).
func (c *Client) Contacts(ctx context.Context) ([]Contact, error) {
	var out struct {
		Contacts []Contact `json:"contacts"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/v1/contacts", url.Values{"limit": {"5000"}}, nil, &out); err != nil {
		return nil, err
	}
	return out.Contacts, nil
}

// ---------------- Pocket Notes ----------------

// Note is a created Pocket Notes note.
type Note struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

// CreateNote writes a markdown note into a folder (category).
func (c *Client) CreateNote(ctx context.Context, title, folder, content string) (*Note, error) {
	var out Note
	body := map[string]any{"title": title, "category": folder, "content": content}
	if err := c.do(ctx, http.MethodPost, "/api/v1/notes", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------------- Pocket Health ----------------

// DayValue is one day of a Pocket Health metric.
type DayValue struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// Metric reads a metric's daily values for the days up to `to` (range: w, m, 3m, 6m, y).
func (c *Client) Metric(ctx context.Context, key, rng, to string) ([]DayValue, error) {
	var out struct {
		Points []DayValue `json:"points"`
	}
	q := url.Values{"range": {rng}}
	if to != "" {
		q.Set("to", to)
	}
	if err := c.do(ctx, http.MethodGet, "/api/v1/metrics/"+url.PathEscape(key), q, nil, &out); err != nil {
		return nil, err
	}
	return out.Points, nil
}
