package rssfetcher

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	http *http.Client
}

type params struct {
	headers map[string]string
	query   map[string]string
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (c *Client) do(ctx context.Context, method string, url string, pms params, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("Error reqest creation: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	for k, v := range pms.headers {
		req.Header.Set(k, v)
	}

	query := req.URL.Query()
	for k, v := range pms.query {
		query.Set(k, v)
	}
	req.URL.RawQuery = query.Encode()

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("Error reqest sending: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("Error response from server: statusCode: %d, body: %s", res.StatusCode, body)
	}

	if err = xml.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("Error decoder: %w", err)
	}
	return nil
}
