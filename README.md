# MangaScraper

A self-hosted manga downloader in Go, utilizing Flaresolverr to solve Cloudflare challenge and scrape mangas off aggregator sites. Example web app is also attached with the docker compose setup as a flutter PWA, accessible on mobile browsers. 
Inspiration is to utilize tailscale to have a setup that allows users (me) to read manga from wherever I want from my home PC.

## Dependencies

- **Tailscale network connection** — required for accessing the services remotely
- **Docker & Docker Compose** — for running the containerized services
- **PostgreSQL** — database running on the host (or accessible from Docker)
- **`psql`** — PostgreSQL client for database setup
- **`goose`** — Go migration tool (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

## Installation

### 1. Clone the repository

```bash
git clone <repo-url>
cd mangascraper
```

### 2. Configure environment variables

Copy the example file and fill in your values:

```bash
cp .env.example .env
```

Edit `.env`:

```env
# Database (point to your PostgreSQL instance)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-db-password
DB_NAME=mangascraper

# Tailscale machine name (or IP/hostname)
API_HOST=your-tailscale-machine-name
IMAGE_HOST=your-tailscale-machine-name
```

### 3. Create the database

```bash
make db-create
```

### 4. Run database migrations

```bash
make migrate-up
```

### 5. Create the image storage directory

```bash
bash scripts/setup-nginx-image-server.sh
```

This creates `/data/mangascraper` for downloaded images.

### 6. Start all services

```bash
docker compose up -d --build
```

This starts:
- **FlareSolverr** on port `8191`
- **Scraper API** on port `8081`
- **Image server** (nginx) on port `8080`
- **Web app** (Flutter) on port `8082`

### 7. Access the app

Open your browser to:

```
http://<your-tailscale-machine-name>:8082
```

## Useful Commands

| Command | Description |
|---------|-------------|
| `make db-create` | Create the PostgreSQL database |
| `make db-drop` | Drop the PostgreSQL database |
| `make migrate-up` | Run all pending migrations |
| `make migrate-down` | Rollback one migration |
| `make migrate-status` | Check migration status |
| `make reset` | Drop, recreate, and migrate the database |
| `docker compose up -d --build` | Build and start all services |
| `docker compose logs -f` | Follow service logs |

## Project Structure

```
├── api/           # Go backend API
├── app/           # Flutter web frontend
├── scripts/       # Setup scripts
├── docker-compose.yml
├── Makefile       # Database migration command
```
