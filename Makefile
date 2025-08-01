# File: Makefile

APP_NAME = social-network
COMPOSE = docker-compose

.PHONY: up down build clean logs

up:
	$(COMPOSE) up --build

down:
	$(COMPOSE) down

build:
	$(COMPOSE) build --no-cache

logs:
	$(COMPOSE) logs -f

clean:
	$(COMPOSE) down --volumes --remove-orphans
	docker rmi $$(docker images -q --filter=reference='$(APP_NAME)-*') || true
