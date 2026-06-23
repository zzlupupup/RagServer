# RagServer

RagServer is a Go + React RAG MCP service with open registration, JWT management APIs, teacher/student roles, local file persistence, MySQL metadata, Redis Stack vector search, and Eino-based ingestion/retrieval.

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
docker compose -f deploy/docker-compose.yml up -d mysql redis-stack
```

Run backend:

```powershell
cd backend
$env:JWT_SECRET="dev-jwt-secret-change-me"
$env:API_KEY_ENCRYPTION_SECRET="change-me-in-production"
$env:EMBEDDING_PROVIDER="openai"
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

## Auth And Roles

- Users register openly as `teacher` or `student`.
- REST management APIs use `Authorization: Bearer <jwt>`.
- Teachers create public knowledge bases visible to all users.
- Students create private knowledge bases visible only to themselves.
- Only teachers can create and manage MCP API keys.
- A teacher-issued API key is bound to a user; MCP requests execute as that bound user.

## MCP

Endpoint:

```text
/mcp
```

Authorization:

```text
Authorization: Bearer <api_key>
```

Tools:

- `kb.list`
- `kb.upload_file`
- `rag.search`

MCP uploads are two-step. First upload the local file to the unauthenticated temporary endpoint:

```bash
curl -F "file=@/path/to/local.pdf" http://host/api/v1/mcp/files/upload
```

Then call `kb.upload_file` with the returned `file_path` and target `kb_name`.

Example tool arguments:

```json
{
  "kb_name": "Course Materials",
  "file_path": "storage/tmp/mcp/2026-06-22/upload_xxx.pdf"
}
```

The temporary upload endpoint only stores a file under `storage/tmp/mcp`; it does not create database records or write Redis indexes. The MCP tool performs the actual authorization and import.

## Required Environment

```text
MYSQL_DSN
REDIS_ADDR
FILE_STORAGE_DIR
MCP_TMP_DIR
OPENAI_BASE_URL
OPENAI_API_KEY
ARK_BASE_URL
ARK_API_KEY
EMBEDDING_PROVIDER
EMBEDDING_MODEL
EMBEDDING_DIMENSION
INDEX_TIMEOUT_SECONDS
JWT_SECRET
JWT_EXPIRES_HOURS
API_KEY_ENCRYPTION_SECRET
MCP_UPLOAD_MAX_MB
```

## Embedding Providers

OpenAI-compatible mode:

```text
EMBEDDING_PROVIDER=openai
OPENAI_BASE_URL=https://api.openai.com
OPENAI_API_KEY=...
EMBEDDING_MODEL=text-embedding-3-small
```

Volcengine Ark multimodal mode:

```text
EMBEDDING_PROVIDER=ark_multimodal
ARK_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
ARK_API_KEY=...
EMBEDDING_MODEL=ep-20260611205227-krftn
```

The current RAG pipeline indexes parsed document text, so Ark multimodal requests are sent as text parts:

```json
{
  "model": "ep-20260611205227-krftn",
  "input": [
    {
      "type": "text",
      "text": "chunk text"
    }
  ]
}
```

## PDF Text Extraction

The first version supports text-layer PDFs only:

```text
PDF -> Eino PDF parser -> pdftotext fallback -> chunks -> embedding
```

Scanned/image-only PDFs still require OCR and will fail with `pdf has no extractable text`.
The backend Docker image installs `poppler-utils` so `pdftotext` is available in containers.
