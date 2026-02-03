package rssfetcher

import (
	"context"
	"html"
	"net/http"
)

func (c *Client) FetchFeed(ctx context.Context, url string) (*RSSFeed, error) {
	var result RSSFeed
	err := c.do(ctx, http.MethodGet, url, params{
		headers: map[string]string{
			"Accept":     "application/xml",
			"User-Agent": "gator",
		},
	}, &result)
	if err != nil {
		return &result, err
	}
	result.Channel.Title = html.UnescapeString(result.Channel.Title)
	result.Channel.Description = html.UnescapeString(result.Channel.Description)
	for _, i := range result.Channel.Item {
		i.Title = html.UnescapeString(i.Title)
		i.Description = html.UnescapeString(i.Description)
	}
	return &result, nil
}

func New() *Client {
	c := &Client{
		http: &http.Client{},
	}
	return c
}
