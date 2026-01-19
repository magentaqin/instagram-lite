SHELL := /bin/bash

.PHONY: help install dev frontend-install frontend-dev backend-install backend-dev clean

help:
	@echo "Targets:"
	@echo "  make install           Install frontend + backend deps"
	@echo "  make dev               Run frontend dev server + backend (concurrently)"
	@echo "  make frontend-install  npm install in ./frontend"
	@echo "  make frontend-dev      npm run dev in ./frontend"
	@echo "  make backend-install   go mod download"
	@echo "  make backend-dev       go run main.go"
	@echo "  make clean             Remove frontend node_modules"

install: frontend-install backend-install

frontend-install:
	cd frontend && npm install && cd ..

backend-install:
	cd backend && go mod download && cd ..

frontend-dev:
	cd frontend && npm run dev && cd ..

backend-dev:
	cd backend && go run main.go && cd ..

dev:
	@echo "Starting frontend + backend..."
	@trap 'kill 0' SIGINT SIGTERM EXIT; \
	$(MAKE) frontend-dev & \
	$(MAKE) backend-dev & \
	wait

clean:
	rm -rf frontend/node_modules