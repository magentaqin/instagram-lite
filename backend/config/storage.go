package config

import (
	"context"
	"log"
	"os"

	"instagram-lite-backend/internal/storage"
)

var Uploader *storage.SpacesUploader

func InitStorage() {
	log.Println(">>> InitStorage called")
	cfg := storage.SpacesConfig{
		Endpoint:      os.Getenv("S3_ENDPOINT"),
		Region:        os.Getenv("S3_REGION"),
		Bucket:        os.Getenv("S3_BUCKET"),
		AccessKey:     os.Getenv("S3_ACCESS_KEY"),
		SecretKey:     os.Getenv("S3_SECRET_KEY"),
		PublicBaseURL: os.Getenv("S3_PUBLIC_URL"),
	}

	log.Printf("S3 config: endpoint=%q region=%q bucket=%q accessKeySet=%v secretKeySet=%v publicBaseURL=%q",
		cfg.Endpoint, cfg.Region, cfg.Bucket,
		cfg.AccessKey != "", cfg.SecretKey != "", cfg.PublicBaseURL,
	)

	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		log.Println("Warning: S3 storage not configured. Upload functionality will be disabled.")
		return
	}

	var err error
	Uploader, err = storage.NewSpacesUploader(context.Background(), cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize storage: %v", err)
		return
	}

	log.Println("Storage initialized successfully")
}
