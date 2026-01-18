package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type PostsHandler struct {
	db *sql.DB
}

func NewPostsHandler(db *sql.DB) *PostsHandler {
	return &PostsHandler{db: db}
}

type CreatePostRequest struct {
	ImageURL string   `json:"image_url"`
	Title    string   `json:"title"`
	Tags     []string `json:"tags"`
}

type PostResponse struct {
	ID        int64    `json:"id"`
	ImageURL  string   `json:"image_url"`
	Title     string   `json:"title"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
}

// Handler for create post
func (h *PostsHandler) CreatePost(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.ImageURL = strings.TrimSpace(req.ImageURL)

	if req.ImageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image_url is required"})
		return
	}
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	// Normalize tags: trim, lowercase, bytes limit, dedupe, 
	tags := normalizeTags(req.Tags)
	if len(tags) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too many tags (max 10)"})
		return
	}

	post, err := h.createPostTx(c, req.ImageURL, req.Title, tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create post failed"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// Create Post Transaction
func (h *PostsHandler) createPostTx(c *gin.Context, imageURL, title string, tags []string) (*PostResponse, error) {
	// start transaction
	tx, err := h.db.BeginTx(c.Request.Context(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	// rollback if not committed
	committed := false
	defer func() {
  if !committed {
    _ = tx.Rollback()
  }
}()

	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	// 1) Insert post
	res, err := tx.ExecContext(
		c.Request.Context(),
		`INSERT INTO posts (image_url, title, created_at) VALUES (?, ?, ?)`,
		imageURL, title, createdAt,
	)
	if err != nil {
		return nil, err
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 2) Upsert tags + 3) join
	for _, t := range tags {
		// tags.name should be unique.
		if _, err := tx.ExecContext(
			c.Request.Context(),
			`INSERT INTO tags (name) VALUES (?) ON CONFLICT(name) DO NOTHING`,
			t,
		); err != nil {
			return nil, err
		}

		var tagID int64
		if err := tx.QueryRowContext(
			c.Request.Context(),
			`SELECT id FROM tags WHERE name = ?`,
			t,
		).Scan(&tagID); err != nil {
			return nil, err
		}

		// post_tags also should have UNIQUE(post_id, tag_id)
		// "ON CONFLICT DO NOTHING" is used to avoid duplicate insertions (such as duplicates in tags or duplicate requests).
		if _, err := tx.ExecContext(
			c.Request.Context(),
			`INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?) ON CONFLICT(post_id, tag_id) DO NOTHING`,
			postID, tagID,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	committed = true

	return &PostResponse{
		ID:        postID,
		ImageURL:  imageURL,
		Title:     title,
		Tags:      tags,
		CreatedAt: createdAt,
	}, nil
}

func normalizeTags(input []string) []string {
	// create a map（use struct{} as the value to avoid extra allocations）
	inputMap := make(map[string]struct{}, len(input))
	output := make([]string, 0, len(input))

	for _, raw := range input {
		t := strings.ToLower(strings.TrimSpace(raw))
		// if the raw tag is empty string
		if t == "" {
			continue
		}
		// limit the tags bytes(16 bytes)
		if len(t) > 16 {
			t = t[:16]
		}
		// dedupe
		if _, ok := inputMap[t]; ok {
			continue
		}
		inputMap[t] = struct{}{}
		output = append(output, t)
	}
	return output
}

// optional helper if you want to distinguish some DB errors later
func isConstraintErr(err error) bool {
	return err != nil && (errors.Is(err, sql.ErrNoRows))
}