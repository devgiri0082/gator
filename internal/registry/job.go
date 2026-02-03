package registry

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/devgiri0082/gator/internal/database"
	rssfetcher "github.com/devgiri0082/gator/internal/rssFetcher"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (r *Registry) fetcherJob(ctx context.Context, d time.Duration) {
	ticker := time.NewTicker(d)
	for ; ; <-ticker.C {
		feeds, err := r.scrapeFeeds(ctx)
		if err != nil {
			fmt.Print(err)
			break
		}
		for key, feed := range feeds {
			curFeed, err := r.db.GetFeed(ctx, sql.NullString{String: key, Valid: true})
			count := 0
			if err != nil {
				fmt.Printf("Unable to get feed: %f\n", err)
				continue
			}
			for _, post := range feed.Channel.Item {
				pubDate, err := parseTime(post.PubDate)
				if err != nil {
					fmt.Printf("unable to parse date: %s", err)
				}
				newPost, err := r.db.CreatePost(ctx, database.CreatePostParams{
					Title:  post.Title,
					Url:    post.Link,
					FeedID: uuid.NullUUID{UUID: curFeed.ID, Valid: true},
					PublishedAt: sql.NullTime{
						Time: pubDate,
					},
					Description: sql.NullString{
						String: post.Description,
						Valid:  true,
					},
				})
				if err != nil {
					var pgErr *pq.Error
					if errors.As(err, &pgErr) && pgErr.Code == "23505" {
						// fmt.Printf("duplicate post: %s\n", post.Link)
						continue
					}
					fmt.Printf("unable to create post: %f\n", err)
				}
				fmt.Printf("Successfully created id: %s,  post: title: %s, url: %s\n", newPost.ID, newPost.Title, newPost.Url)
				count++
			}
			fmt.Printf("successfully added: %d new posts from url: %s\n", count, key)
		}
		fmt.Println("Complted running this batch")
	}
}

func parseTime(date string) (time.Time, error) {
	var timeLayouts = []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
	}
	for _, layout := range timeLayouts {
		pubDate, err := time.Parse(layout, date)
		if err == nil {
			return pubDate, nil
		}
	}
	return time.Time{}, errors.New("Invalid format")
}

func (r *Registry) scrapeFeeds(ctx context.Context) (map[string]*rssfetcher.RSSFeed, error) {
	feeds, err := r.db.GetNextFeedToFetch(ctx)
	var fetchedFeeds map[string]*rssfetcher.RSSFeed = map[string]*rssfetcher.RSSFeed{}
	if err != nil {
		return fetchedFeeds, fmt.Errorf("Error fetching feeds data: %w", err)
	}
	for _, feed := range feeds {
		f, err := r.client.FetchFeed(ctx, feed.Url.String)
		if err != nil {
			fmt.Printf("unable to fetch feed: %s, err: %f", feed.Url.String, err)
			continue
		}
		fetchedFeeds[feed.Url.String] = f
		err = r.db.MarkFeedFetched(ctx, feed.ID)
		if err != nil {
			fmt.Printf("Unable to update last fetch: %s", err)
		}
	}
	return fetchedFeeds, nil
}
