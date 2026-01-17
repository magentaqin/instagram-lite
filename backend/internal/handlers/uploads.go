package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"

	"instagram-lite-backend/internal/imageproc"
	"instagram-lite-backend/internal/storage"
)

type UploadHandler struct {
	uploader *storage.SpacesUploader
}

func NewUploadHandler(u *storage.SpacesUploader) *UploadHandler {
	return &UploadHandler{uploader: u}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	ctx := c.Request.Context()

	// Limit body size
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, imageproc.MaxUploadBytes)

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Get the uploaded file stream so we can read and process its bytes.
	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open file"})
		return
	}
	// ensure file descriptor is released
	defer f.Close()

	// resize and crop the image to 512 * 512
	// TODO: move image processing to a background worker if upload throughput becomes a bottleneck
	jpegBytes, err := imageproc.ProcessJPEG(f)
	if err != nil {
		if err.Error() == "file too large" {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large"})
			return
		}
		if err.Error() == "decode failed" || err.Error() == "encode failed" {
			c.JSON(http.StatusBadGateway, gin.H{"error": "image encoding/decoding failed"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate image key and upload it to S3
	uploadID := ulid.Make().String()
	key := fmt.Sprintf("uploads/%s.jpg", uploadID)

	publicURL, err := h.uploader.PutJPEG(ctx, key, jpegBytes)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "upload image failed"})
		return
	}

	// return image URL
	c.JSON(http.StatusCreated, gin.H{"image_url": publicURL})
}
