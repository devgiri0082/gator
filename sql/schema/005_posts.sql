-- +goose Up
CREATE TABLE posts (
  id UUID PRIMARY KEY,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now(),
  title TEXT NOT NULL,
  url TEXT NOT NULL UNIQUE,
  description TEXT,
  published_at TIMESTAMPTZ DEFAULT now(),
  feed_id UUID REFERENCES feeds(id)
);

-- +goose Down
DROP TABLE posts;
