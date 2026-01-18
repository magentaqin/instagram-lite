PRAGMA foreign_keys = ON;

-- Posts
-- id: internal primary key 
-- post_id: public id 
CREATE TABLE IF NOT EXISTS posts (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  post_id    TEXT    NOT NULL UNIQUE,
  title      TEXT    NOT NULL,
  image_url  TEXT    NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_posts_created_at
  ON posts(created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_posts_post_id
  ON posts(post_id);

-- Tags
CREATE TABLE IF NOT EXISTS tags (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  name       TEXT    NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tags_name
  ON tags(name);

-- Join table (N to N)
-- post_db_id references posts.id (internal primary key)
CREATE TABLE IF NOT EXISTS post_tags (
  post_db_id INTEGER NOT NULL,
  tag_id     INTEGER NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (post_db_id, tag_id),
  FOREIGN KEY (post_db_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (tag_id)     REFERENCES tags(id)  ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_post_tags_tag_id
  ON post_tags(tag_id, post_db_id);