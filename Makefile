# Makefile for Trae - Ticket Reservation and Event System

.PHONY: help build up down logs ps clean dev test

help:
	@echo "Available commands:"
	@echo "  make build    - Build all services"
	@echo "  make up       - Start all services"
	@echo "  make down     - Stop all services"
	@echo "  make logs     - View logs from all services"
	@echo "  make ps       - List running services"
	@echo "  make clean    - Remove all containers, networks, and volumes"
	@echo "  make dev      - Start development environment"
	@echo "  make test     - Run tests in Docker"

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

ps:
	docker-compose ps

clean:
	docker-compose down -v --remove-orphans

dev:
	docker-compose -f docker-compose.dev.yml up -d

test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	echo "Cleaning up test containers..."
	docker-compose -f docker-compose.test.yml down