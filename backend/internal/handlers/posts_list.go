package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ListPostsResponse struct {
	Items      []PostItem `json:"items"`
	NextCursor *string    `json:"next_cursor"`
	HasMore    bool       `json:"has_more"`
}

type PostItem struct {
	ID        string   `json:"id"`        // public id 
	Title     string   `json:"title"`
	ImageURL  string   `json:"image_url"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"` 
}

type postsCursor struct {
	CreatedAt string `json:"created_at"`
	DBID      int64  `json:"db_id"`
}

func encodeCursor(c postsCursor) (string, error) {
	// Serialize cursor struct into JSON bytes before base64 encoding.
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	// Encode JSON cursor as base64 so it can be passed as an opaque query parameter.
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeCursor(s string) (*postsCursor, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	// decode base64
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	// deserialize 
	var c postsCursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	if c.CreatedAt == "" || c.DBID <= 0 {
		return nil, sql.ErrNoRows // treat it as invalid cursor
	}
	return &c, nil
}

func splitCSVTags(csv sql.NullString) []string {
	if !csv.Valid || strings.TrimSpace(csv.String) == "" {
		return []string{}
	}
	parts := strings.Split(csv.String, ",")
	output := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t == "" {
			continue
		}
		// keep response stable (dedupe just in case)
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		output = append(output, t)
	}
	return output
}

// Valite limit
func parseLimit(c *gin.Context) (int, error) {
	const (
		defaultLimit = 20
		maxLimit     = 50
	)
	s := strings.TrimSpace(c.Query("limit"))
	if s == "" {
		return defaultLimit, nil
	}
	// turn string into integer
	n, err := strconv.Atoi(s)
	if err != nil {
    return 0, err
	}
	if n <= 0 {
		return 0, fmt.Errorf("limit must be > 0")
	}
	if n > maxLimit {
		n = maxLimit
	}
	return n, nil
}

// ListPosts handler
func (h *PostsHandler) ListPosts(c *gin.Context) {
	// Validate Limit
	limit, err := parseLimit(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	// trim and turn it to lower case
	tag := strings.ToLower(strings.TrimSpace(c.Query("tag")))
	// trim cursor 
	cursorStr := strings.TrimSpace(c.Query("cursor"))

	cur, err := decodeCursor(cursorStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor"})
		return
	}

	// Fetch limit+1 to know if there is more.
	limitPlusOne := limit + 1

	// List posts with keyset pagination (created_at, id) and fuzzy tag filter.
	//  Tags are aggregated via GROUP_CONCAT.
	// LEFT JOIN is used to keep posts that have no tags.
	/**
	-- Optional fuzzy tag filter:
  -- If tag query is empty, do not filter by tags.
  -- Otherwise keep the post if it has at least one tag whose name contains the query.
	**/
	/**
	 -- Optional pagination (created_at, id):
   -- If cursor is empty, return the first page.
   -- Otherwise return posts  older than the cursor.
	**/
	q := `
SELECT
  p.id,
  p.post_id,
  p.title,
  p.image_url,
  p.created_at,
  GROUP_CONCAT(t.name) AS tags_csv 
FROM posts p
LEFT JOIN post_tags pt ON pt.post_db_id = p.id
LEFT JOIN tags t ON t.id = pt.tag_id
WHERE
  (
    ? = '' OR
    EXISTS (
      SELECT 1
      FROM post_tags pt2
      JOIN tags t2 ON t2.id = pt2.tag_id
      WHERE pt2.post_db_id = p.id
        AND t2.name LIKE '%' || ? || '%'
    )
  )
  AND (
    ? = '' OR
    (p.created_at < ? OR (p.created_at = ? AND p.id < ?))
  )
GROUP BY p.id
ORDER BY p.created_at DESC, p.id DESC
LIMIT ?;
`

	// Cursor params: if no cursor, we should pass '' to skip.
	curFlag := ""
	curCreatedAt := ""
	curID := int64(0)
	if cur != nil {
		curFlag = "1"
		curCreatedAt = cur.CreatedAt
		curID = cur.DBID
	}

	ctx := c.Request.Context()
	rows, err := h.db.QueryContext(
		ctx,
		q,
		tag, tag,
		curFlag, curCreatedAt, curCreatedAt, curID,
		limitPlusOne,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list posts failed"})
		return
	}
	// Always close rows to release the underlying DB connection.
	defer rows.Close()

	// rowItem is an internal scan target that mirrors the SELECT list.
	type rowItem struct {
		DBID      int64
		PostID    string
		Title     string
		ImageURL  string
		CreatedAt string
		TagsCSV   sql.NullString
	}
	// the final output items
	output := make([]PostItem, 0, limitPlusOne)
	// it stores the scanned DB rows(we need to process them: split tags, compute next cursor, has_more)
	raw := make([]rowItem, 0, limitPlusOne)

	// iterate rows and fill the `raw` 
	for rows.Next() {
		var r rowItem
		// Scan the current row into rowItem
		if err := rows.Scan(&r.DBID, &r.PostID, &r.Title, &r.ImageURL, &r.CreatedAt, &r.TagsCSV); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list posts failed"})
			return
		}
		raw = append(raw, r)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list posts failed"})
		return
	}

  // Previously, we already fetched limit+1 rows.
  // If we got more than limit, mark hasMore and trim back to limit.
	hasMore := false
	if len(raw) > limit {
		hasMore = true
		raw = raw[:limit]
	}

	// raw DBID doesn't need to output to client.
	for _, r := range raw {
		output = append(output, PostItem{
			ID:        r.PostID,
			Title:     r.Title,
			ImageURL:  r.ImageURL,
			Tags:      splitCSVTags(r.TagsCSV),
			CreatedAt: r.CreatedAt,
		})
	}

	// Calcuate next cursor
	var nextCursor *string
	if hasMore && len(raw) > 0 {
		last := raw[len(raw)-1]
		curStr, err := encodeCursor(postsCursor{
			CreatedAt: last.CreatedAt,
			DBID:      last.DBID,
		})
		if err == nil {
			nextCursor = &curStr
		}
	}

	c.JSON(http.StatusOK, ListPostsResponse{
		Items:      output,
		NextCursor: nextCursor, // only returned when hasMore is true.
		HasMore:    hasMore,
	})
}