.DEFAULT_GOAL := help

.PHONY: help build up down clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

build: ## Build docker compose images
	docker compose build

up: ## Start services in the background
	docker compose up -d

down: ## Stop services
	docker compose down

clean: ## Stop and remove containers, volumes, and built images
	docker compose down --rmi local --volumes --remove-orphans
