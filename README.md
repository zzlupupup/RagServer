# RagServer

RagServer is a single-tenant, multi-knowledge-base RAG MCP service.

## Stack

- Backend: Go, Gin, Eino, official MCP Go SDK
- Frontend: React, Vite, TypeScript
- Persistence: MySQL
- Vector search: Redis Stack
- File storage: local filesystem
- Embeddings: OpenAI-compatible `/v1/embeddings`

## Development

Start MySQL and Redis Stack:

```powershell
docker compose -f deploy/docker-compose.yml up -d
```

Run backend:

```powershell
cd backend
$env:ADMIN_TOKEN="dev-admin-token"
$env:API_KEY_ENCRYPTION_SECRET="change-me-in-production"
$env:OPENAI_API_KEY="..."
go run ./cmd/server
```

Run frontend:

```powershell
cd frontend
npm install
npm run dev
```

The frontend uses Vite proxy for `/api` and `/mcp`; the Go backend does not enable CORS.

## MCP Tools

- `kb.list`
- `kb.upload_file`
- `rag.search`

MCP endpoint:

```text
/mcp
```

Use `Authorization: Bearer <api_key>` generated in the knowledge-base management UI.

