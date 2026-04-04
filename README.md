# Go Bank API

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)

A robust, fully Dockerized REST API built in Go for handling core banking operations. This backend system handles accounts, transfers, JWT authentication, and utilizes raw PostgreSQL queries for high-performance data management.

## Project Context

> **Architectural Sandbox:** This repository serves as a foundational architecture testbed. It was designed to establish scalable patterns in Go (Clean Architecture, Interface usage, Dockerization) before integrating these identical structures into my flagship high-concurrency project, **UniGo** (a massive, high-concurrency spatial-temporal ride-pooling system).

## Features

- **Account Management**: Create, read, and delete bank accounts securely.
- **Transfers**: Transfer funds between accounts.
- **Authentication**: JWT-secured endpoints to ensure authorized access.
- **Database**: Direct PostgreSQL integration utilizing raw SQL queries for optimized performance.
- **Dockerized**: Containerized application and database for seamless deployment.

## Tech Stack

- **Language**: [Go](https://golang.org/)
- **Router**: [Gorilla Mux](https://github.com/gorilla/mux)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **Infrastructure**: [Docker](https://www.docker.com/) & Docker Compose

## Getting Started

1. Clone the repository.
2. Run database and application using Docker Compose:
   ```bash
   docker-compose up --build
   ```
3. The API will be accessible on port configured in the Docker setup (default `:3000`).

## Architecture

This project is built focusing on clean architecture principles:
- Robust Interface usage for dependency injection and easy testing.
- Separation of concerns between routing, application logic, and storage.
- Standardized error handling and JSON responses.
