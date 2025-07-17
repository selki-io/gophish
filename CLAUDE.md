# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gophish is an open-source phishing framework written in Go for security awareness training and penetration testing. It consists of:
- Go backend API server (port 3333) for administration
- Phishing campaign server (port 80/443) for landing pages
- SQLite/MySQL database for data persistence
- Web frontend using jQuery, Bootstrap, and modern JavaScript

## Essential Commands

### Building
```bash
# Backend
go build -v .

# Frontend assets (required before running)
npm install --only=dev
gulp
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./models
go test ./controllers/api

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v ./models -check.f ModelsSuite.TestPostGroup
```

### Development
```bash
# Format code (required for PRs)
go fmt ./...

# Check formatting
diff -u <(echo -n) <(gofmt -d .)

# Run server (after building frontend)
./gophish

# Run specific gulp tasks
gulp vendorjs  # Build vendor JavaScript
gulp scripts   # Build application JavaScript
gulp styles    # Build CSS

# Database migrations (if using custom database)
# SQLite: migrations in db/db_sqlite3/migrations/
# MySQL: migrations in db/db_mysql/migrations/
```

## Architecture Overview

### Core Components
- **controllers/api/** - RESTful API endpoints for campaigns, templates, users, groups
- **models/** - Data models and database operations using GORM
- **worker/** - Background job processing for sending emails and managing campaigns
- **mailer/** - Email sending functionality with SMTP support
- **auth/** - Authentication and RBAC implementation
- **webhook/** - Webhook integrations for campaign events
- **middleware/** - HTTP middleware for rate limiting, session management
- **imap/** - IMAP monitoring for bounce handling and email replies
- **logger/** - Structured logging with logrus

### Database
- Uses GORM for database abstraction
- Migrations in `db/db_sqlite3/` and `db/db_mysql/`
- Models define schema and business logic
- Test suite uses in-memory SQLite

### Frontend
- Traditional server-rendered application with jQuery for interactivity
- Build system uses Gulp for main assets, Webpack for specific modules (passwords, users, webhooks)
- JavaScript source in `static/js/src/`, built to `static/js/dist/`
- Templates in `templates/` directory (Go html/template)
- Key libraries: jQuery, Bootstrap, DataTables, Highcharts, D3.js, SweetAlert2

### Security Architecture
- CSRF protection on all state-changing operations
- Rate limiting middleware
- Password policies and account locking
- TLS/SSL support for admin interface
- Separate servers for admin and phishing campaigns

## Testing Approach
- Suite-based testing using gocheck framework (`gopkg.in/check.v1`)
- Each component has corresponding `*_test.go` file
- Tests use test fixtures and mock data
- API tests include full request/response cycle
- Always run tests before committing changes
- CI runs tests on Go 1.21, 1.22, and 1.23

## Key Dependencies
- **gorilla/mux** - HTTP routing
- **gorilla/csrf** - CSRF protection
- **jinzhu/gorm** - ORM for database operations
- **sirupsen/logrus** - Structured logging
- **jordan-wright/email** & **gomail** - Email functionality
- **oschwald/maxminddb-golang** - GeoIP lookups