-- +goose Up
create table feed_follows (
  id UUID PRIMARY KEY,
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  user_id UUID REFERENCES users (id) ON DELETE CASCADE,
  feed_id UUID REFERENCES feeds (id) ON DELETE CASCADE,

  UNIQUE (user_id, feed_id)
);

-- +goose Down
drop table feed_follows;
