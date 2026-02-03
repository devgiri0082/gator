# RSS Fetcher In GO (GATOR)

## Installation

### Requirements

- PostgreSQL database
- Go 1.25.4 or higher

### Command

- Install

  ```bash
  go install github.com/devgiri0082/gator
  ```

- Config
  Create a JSON configuration file at: `~/.gatorconfig.json`

  ```json
  {
    "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
    "current_user_name": ""
  }
  ```

- Usage
  ```bash
  gator <command> [args]
  ```

## Available Commands

### User Management

- `gator register <username>` - Register a new user
- `gator login <username>` - Login as an existing user
- `gator users` - List all registered users
- `gator reset` - Delete all users from database

### Feed Management

- `gator addfeed <name> <url>` - Add a new RSS feed (requires login)
- `gator feeds` - List all available feeds
- `gator follow <url>` - Follow an RSS feed (requires login)
- `gator following` - List feeds you're following (requires login)
- `gator unfollow <url>` - Unfollow an RSS feed (requires login)

### Content

- `gator agg <duration>` - Start fetching feeds at specified interval (e.g., "1m", "30s", "1h")
- `gator browse [limit]` - Browse latest posts from followed feeds (requires login, default limit: 2)

## Example Usage

```bash
# Register and login
gator register john
gator login john

# Add and follow feeds
gator addfeed techcrunch https://techcrunch.com/feed/
gator follow https://techcrunch.com/feed/

# Start aggregating feeds every 60 seconds
gator agg 60s

# Browse latest posts
gator browse 10
```
