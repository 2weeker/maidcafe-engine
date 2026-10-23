COMPOSE := docker compose

.DEFAULT_GOAL := help
.PHONY: help up down build rebuild restart logs ps psql adduser generate fmt vet clean

help: ## list targets
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-10s %s\n", $$1, $$2}'

up: ## build and start everything in the background
	$(COMPOSE) up -d --build

down: ## stop everything
	$(COMPOSE) down

build: ## build the app image
	$(COMPOSE) build app

rebuild: ## rebuild and restart the app (starts the db too if it's missing)
	$(COMPOSE) up -d --build app

restart: ## restart containers without rebuilding
	$(COMPOSE) restart

logs: ## follow app logs
	$(COMPOSE) logs -f app

ps: ## show container status
	$(COMPOSE) ps

psql: ## open a psql shell on the database
	$(COMPOSE) exec db sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"'

adduser: ## create an account: make adduser NAME=alice [ROLE=mod] (asks for the password)
	$(COMPOSE) run --rm --no-deps app adduser $(NAME) $(ROLE)

generate: ## regenerate templ files locally (for editor support)
	go tool templ generate

fmt: ## format go and templ sources
	go tool templ fmt web/templates
	gofmt -w app cmd web

vet: generate ## static checks
	go vet ./...

clean: ## stop everything and DELETE the database volume
	$(COMPOSE) down -v
