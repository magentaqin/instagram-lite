package imageproc

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"net/http"

	"github.com/disintegration/imaging"
)

const (
	MaxUploadBytes = int64(10 << 20) // 10MB
)

func ProcessJPEG(r io.Reader) ([]byte, error) {
	// Check size
	// Although we have http.MaxBytesReader in handler, this limit reader here is to keep this function safe without depending on the handler. e.g. Unit Test
	lr := &io.LimitedReader{R: r, N: MaxUploadBytes + 1}
	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > MaxUploadBytes {
		return nil, fmt.Errorf("file too large")
	}

	// Check Type. 
	ct := http.DetectContentType(head512(data))
	if ct != "image/jpeg" && ct != "image/png" {
		return nil, fmt.Errorf("unsupported content-type: %s", ct)
	}

	// Convert bytes into image.Image for cropping and resizing.
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode failed")
	}

	// For png image file, we need to normalize its transaparent background.
	//  PNG may contain alpha; JPEG doesn't support transparency.
	if ct == "image/png" {
		img = normalizeBgForPng(img)
	}

	// center crop square + resize
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	// the shortest size will be the square size
	size := min(w, h)
	// crop a square shape from the center.
	square := imaging.CropCenter(img, size, size)
	// resize it to 512 * 512
	out := imaging.Resize(square, 512, 512, imaging.Lanczos)

	// after cropping and resizing, we need to encode it back. 
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, out, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("encode failed")
	}
	return buf.Bytes(), nil
}


// Get the min number
func min(a, b int) int {
  if a < b { return a }
  return b
}


// Get the first 512 bytes to check the image type.
func head512(b []byte) []byte {
	if len(b) <= 512 {
		return b
	}
	return b[:512]
}

// As we'll store the image into jpg format(not support transparency), for png images, it has transparency.
// We need to fill the transparency background with white to make it look normal.
func normalizeBgForPng(img image.Image) image.Image {
	// create a new canvas
	dst := image.NewRGBA(img.Bounds())
	// fill it with white background
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	// put the image over the background color
	draw.Draw(dst, dst.Bounds(), img, img.Bounds().Min, draw.Over)
	return dst
}