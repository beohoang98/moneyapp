# MoneyApp

A web application for managing personal finances — track expenses, incomes, and invoices/bills.

## Tech Stack

- **Frontend**: React (TypeScript)
- **Backend**: Go (REST API)
- **Database**: SQLite (embedded)
- **File Storage**: MinIO (for invoices, receipts, and attachments)

## Project Structure

```
moneyapp/
├── frontend/          # React TypeScript app
├── backend/           # Go API server
├── docker-compose.yml # MinIO and dev services
└── README.md
```

## Features

- Expense tracking and categorization
- Income management
- Invoice and bill management with file attachments
- Financial overview and summaries

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker (for MinIO)

## Getting Started

### Backend

```bash
cd backend
go mod download
go run ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### MinIO (File Storage)

```bash
docker compose up minio -d
```

MinIO console will be available at `http://localhost:9001`.
