-- seed.sql
PRAGMA foreign_keys = ON;

-- 1) Insert tags
INSERT OR IGNORE INTO tags (name) VALUES
('sunset'), ('nature'), ('photography'),
('coffee'), ('morning'), ('cafe'),
('weekend'), ('vibes'), ('relax'),
('outdoors'), ('hiking'),
('friends'), ('fun'), ('memories'),
('adventure'), ('travel'), ('explore'),
('grateful'), ('blessed'), ('life'),
('happiness'), ('simple'), ('moments');

-- 2) Insert 150 posts
WITH RECURSIVE seq(i) AS (
  SELECT 1
  UNION ALL
  SELECT i + 1 FROM seq WHERE i < 150
)
INSERT INTO posts (post_id, title, image_url, created_at)
SELECT
  'mock-' || i AS post_id,
  CASE (i - 1) % 8
    WHEN 0 THEN 'Beautiful sunset today! 🌅'
    WHEN 1 THEN 'Coffee time ☕'
    WHEN 2 THEN 'Weekend vibes ✨'
    WHEN 3 THEN 'Nature is amazing'
    WHEN 4 THEN 'Good times with friends'
    WHEN 5 THEN 'New adventure begins'
    WHEN 6 THEN 'Feeling grateful today'
    ELSE        'Simple moments, big happiness'
  END AS title,
  'https://picsum.photos/seed/' || (i + 99) || '/600/600' AS image_url,
  datetime('now', '-' || ((i - 1) % 30) || ' days') AS created_at
FROM seq;

-- 3) Insert post_tags: 3 tags per post 
WITH RECURSIVE seq(i) AS (
  SELECT 1
  UNION ALL
  SELECT i + 1 FROM seq WHERE i < 150
),
tag_groups(grp, name) AS (
  SELECT 0, 'sunset'      UNION ALL
  SELECT 0, 'nature'      UNION ALL
  SELECT 0, 'photography' UNION ALL

  SELECT 1, 'coffee'      UNION ALL
  SELECT 1, 'morning'     UNION ALL
  SELECT 1, 'cafe'        UNION ALL

  SELECT 2, 'weekend'     UNION ALL
  SELECT 2, 'vibes'       UNION ALL
  SELECT 2, 'relax'       UNION ALL

  SELECT 3, 'nature'      UNION ALL
  SELECT 3, 'outdoors'    UNION ALL
  SELECT 3, 'hiking'      UNION ALL

  SELECT 4, 'friends'     UNION ALL
  SELECT 4, 'fun'         UNION ALL
  SELECT 4, 'memories'    UNION ALL

  SELECT 5, 'adventure'   UNION ALL
  SELECT 5, 'travel'      UNION ALL
  SELECT 5, 'explore'     UNION ALL

  SELECT 6, 'grateful'    UNION ALL
  SELECT 6, 'blessed'     UNION ALL
  SELECT 6, 'life'        UNION ALL

  SELECT 7, 'happiness'   UNION ALL
  SELECT 7, 'simple'      UNION ALL
  SELECT 7, 'moments'
)
INSERT OR IGNORE INTO post_tags (post_db_id, tag_id, created_at)
SELECT
  p.id AS post_db_id,
  t.id AS tag_id,
  p.created_at
FROM seq
JOIN posts p
  ON p.post_id = 'mock-' || seq.i
JOIN tag_groups g
  ON g.grp = ((seq.i - 1) % 8)
JOIN tags t
  ON t.name = g.name;