package registry

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/devgiri0082/gator/internal/config"
	"github.com/devgiri0082/gator/internal/database"
	rssfetcher "github.com/devgiri0082/gator/internal/rssFetcher"
	"github.com/google/uuid"
)

type Registry struct {
	commands map[string]Command
	config   *config.Config
	db       *database.Queries
	client   *rssfetcher.Client
}

type CommandFunc func(ctx context.Context, args []string) error
type Command struct {
	Name        string
	Description string
	Usage       string
	ArgCount    int
	Opt         bool
	callback    CommandFunc
}

func (r *Registry) Run(key string, args []string) error {
	cmd, ok := r.commands[key]
	if !ok {
		return fmt.Errorf("command: %s not found", key)
	}
	if !cmd.Opt && len(args) != cmd.ArgCount {
		return errors.New(cmd.Usage)
	}
	return cmd.callback(context.Background(), args)
}

func (r *Registry) register(cmd Command) {
	r.commands[cmd.Name] = cmd
}

func New(config *config.Config, db *database.Queries, client *rssfetcher.Client) *Registry {
	r := &Registry{
		commands: map[string]Command{},
		config:   config,
		db:       db,
		client:   client,
	}
	r.register(
		Command{
			Name:        "login",
			Description: "login to the application",
			Usage:       "Usage: gator login <username>",
			ArgCount:    1,
			callback: func(ctx context.Context, args []string) error {
				username := args[0]
				user, err := r.db.GetUser(ctx, username)
				if err != nil && errors.Is(err, sql.ErrNoRows) {
					return errors.New("You can't login to an account that doesn't exist!")
				} else if err != nil {
					return err
				}
				err = r.config.SetUser(user.Name)
				if err != nil {
					return err
				}
				fmt.Printf("successfully set user: %s", username)
				return nil
			},
		})

	r.register(Command{
		Name:        "register",
		Description: "register a user",
		Usage:       "Usage gator register <username>",
		ArgCount:    1,
		callback: func(ctx context.Context, args []string) error {
			username := args[0]
			user, err := r.db.CreateUser(ctx, username)
			if err != nil {
				return fmt.Errorf("user creation errror: %w", err)
			}
			r.config.SetUser(user.Name)
			fmt.Printf("Successfully Created User %s\n", user.Name)
			fmt.Println(user)
			return nil
		},
	})

	r.register(Command{
		Name:        "reset",
		Description: "delete users from db",
		Usage:       "usage: gator reset",
		ArgCount:    0,
		callback: func(ctx context.Context, args []string) error {
			err := r.db.Resetusers(ctx)
			if err != nil {
				return err
			}
			fmt.Println("Successfully deleted all the users")
			return nil
		},
	})

	r.register(Command{
		Name:        "users",
		Description: "print all the users",
		Usage:       "usage: gator users",
		ArgCount:    0,
		callback: func(ctx context.Context, args []string) error {
			users, err := r.db.GetUsers(ctx)
			if err != nil {
				return err
			}
			currUser := r.config.Getuser()
			var result strings.Builder
			for _, u := range users {
				if u == currUser {
					fmt.Fprintf(&result, "* %s (current)\n", u)
					continue
				}
				fmt.Fprintf(&result, "* %s\n", u)

			}
			fmt.Print(result.String())
			return nil
		},
	})

	r.register(Command{
		Name:        "agg",
		Description: "fetch the entire feed",
		Usage:       "usage: gator agg <duration>",
		ArgCount:    1,
		callback: func(ctx context.Context, args []string) error {
			dur, err := time.ParseDuration(args[0])
			if err != nil {
				return fmt.Errorf("Invalid duration: %w", err)
			}
			fmt.Printf("Collecting feeds every %s\n", dur)
			r.fetcherJob(ctx, dur)
			return nil
		},
	})

	r.register(Command{
		Name:        "addfeed",
		Description: "add feed to the database",
		Usage:       "usage: gator addfeed <name> <url>",
		ArgCount:    2,
		callback: r.middlewareLoggedIn(func(ctx context.Context, args []string) error {
			name := args[0]
			url := args[1]
			user := UserFromContext(ctx)
			feed, err := r.db.AddFeed(ctx, database.AddFeedParams{Name: sql.NullString{String: name, Valid: true}, Url: sql.NullString{String: url, Valid: true}, UserID: uuid.NullUUID{UUID: user.ID, Valid: true}})
			if err != nil {
				return err
			}

			feedFollow, err := r.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
				UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
				FeedID: uuid.NullUUID{UUID: feed.ID, Valid: true},
			})

			if err != nil {
				return fmt.Errorf("Unable to create feed follow: %w", err)

			}

			fmt.Printf("Successfully created feed: %s and followed by user: %s", feedFollow.FeedName.String, feedFollow.UserName)

			return nil
		}),
	})

	r.register(Command{
		Name:        "feeds",
		Description: "list all the available feeds",
		Usage:       "usage: gator feeds",
		ArgCount:    0,
		callback: r.middlewareLoggedIn(func(ctx context.Context, args []string) error {
			users, err := r.db.GetFeeds(ctx)
			if err != nil {
				return err
			}
			var final strings.Builder
			fmt.Fprint(&final, "Feeds Info: \n")
			for _, u := range users {
				fmt.Fprintf(&final, "name: %s, url: %s, username: %s\n", u.Name.String, u.Url.String, u.Username)
			}
			fmt.Println(final.String())
			return nil
		}),
	})

	r.register(Command{
		Name:        "follow",
		Description: "follow a feed",
		ArgCount:    1,
		Usage:       "usage: gator follow <url>",
		callback: r.middlewareLoggedIn(func(ctx context.Context, args []string) error {
			url := args[0]
			user := UserFromContext(ctx)
			feed, err := r.db.GetFeed(ctx, sql.NullString{String: url, Valid: true})
			if err != nil {
				return fmt.Errorf("Unable to fetch feed: %w", err)
			}
			feedFollow, err := r.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
				UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
				FeedID: uuid.NullUUID{UUID: feed.ID, Valid: true},
			})
			if err != nil {
				return fmt.Errorf("Unable to create feed follow: %w", err)
			}
			fmt.Printf("User : %s, successfully followed the feed: %s\n", feedFollow.UserName, feedFollow.FeedName.String)
			return nil
		}),
	})

	r.register(Command{
		Name:        "following",
		Description: "List following feeds",
		ArgCount:    0,
		Usage:       "usage: gator following",
		callback: r.middlewareLoggedIn(func(ctx context.Context, args []string) error {
			user := UserFromContext(ctx)
			followings, err := r.db.GetFeedFollowsForUser(ctx, uuid.NullUUID{UUID: user.ID, Valid: true})
			if err != nil {
				return fmt.Errorf("unable to fetching followings: %w", err)
			}

			fmt.Printf("User %s is following feeds: \n", user.Name)
			for _, follow := range followings {
				fmt.Printf("\t-%s\n", follow.FeedName.String)
			}
			return nil

		}),
	})

	r.register(Command{
		Name:        "unfollow",
		Description: "unfollow a RSS feed",
		ArgCount:    1,
		Usage:       "usage: gator unfollow <url>",
		callback: r.middlewareLoggedIn(func(ctx context.Context, args []string) error {
			user := UserFromContext(ctx)
			url := args[0]
			feed, err := r.db.GetFeed(ctx, sql.NullString{String: url, Valid: true})
			if err != nil {
				return fmt.Errorf("Unable to get feed: %w", err)
			}
			_, err = r.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
				UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
				FeedID: uuid.NullUUID{UUID: feed.ID, Valid: true},
			})
			if err != nil {
				return fmt.Errorf("Unable to unfollow: %w", err)
			}

			fmt.Printf("Successfully unfollowed feed %s, from user %s", feed.Name.String, user.Name)
			return nil
		}),
	})

	r.register(Command{
		Name:        "browse",
		Description: "browse latest posts",
		ArgCount:    1,
		Opt:         true,
		Usage:       "usage: gator browse <limit:optional>",
		callback: r.middlewareLoggedIn(func(ctx context.Context, args []string) error {
			user := UserFromContext(ctx)
			var limit int32 = 2
			if len(args) == 1 {
				val, err := strconv.Atoi(args[0])
				if err == nil {
					limit = int32(val)
				}
			}
			posts, err := r.db.GetPostFromUser(ctx, database.GetPostFromUserParams{
				UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
				Limit:  int32(limit),
			})
			if err != nil {
				return fmt.Errorf("Unable to fetch posts: %w", err)
			}
			var postStr strings.Builder
			fmt.Fprint(&postStr, "Recent Posts:\n")
			for _, p := range posts {
				fmt.Fprintf(&postStr, "\t-%s\n", p.Title)
			}
			fmt.Println(postStr.String())
			return nil
		}),
	})

	return r
}
