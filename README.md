# Pulse — Live Polling App

Create a poll, share a link, and watch results update for every viewer in
real time — no page refresh.

```
Create Poll → Share Link → Audience Votes → Results Update Live
```

## Architecture

```
React (Vite)  ──HTTP/WS──▶  Go / Gin API
                                 │
                     ┌───────────┴───────────┐
                     ▼                       ▼
                 MongoDB                   Redis
              (source of truth)      (live counts + Pub/Sub)
                                           │
                                    publish on vote
                                           ▼
                                    WebSocket Hub
                                           │
                             ┌─────────────┼─────────────┐
                             ▼             ▼             ▼
                         Client A      Client B      Client C
```

**Realtime path, precisely:**

```
POST /api/polls/:id/vote
   → validate poll is active + option exists
   → Redis SETNX dedupe check (fast path)
   → MongoDB insert vote (unique index on pollId+voterIdentifier = durable dedupe)
   → Redis HINCRBY live counter
   → Redis PUBLISH poll:{id}:updates
   → Hub's single subscription for that poll receives the message
   → Hub fans it out to every WebSocket client currently viewing that poll
   → Every client updates its bars, no refresh
```

Only **one** Redis subscription is opened per poll (shared by every
viewer of that poll), not one per WebSocket connection — see
`internal/websocket/hub.go`.

## Features

- Email/password auth (JWT, bcrypt-hashed passwords)
- Create polls (2–10 options, optional description/expiry/anonymous toggle)
- Public, unauthenticated voting with server-side duplicate-vote prevention
- Live results over WebSocket, backed by Redis Pub/Sub
- Live viewer count per poll
- Poll closing (manual + automatic on expiry)
- Dashboard: view/copy-link/close/delete your polls
- Rate limiting on auth + voting endpoints
- Light/dark theme
- Responsive, accessible UI (semantic HTML, labeled inputs, focus states)

## Tech stack

React 18 (Vite) · React Router · Axios · Go · Gin · MongoDB (official
driver) · Redis (go-redis v9) · gorilla/websocket · JWT · bcrypt · Docker

## Local setup

**Requirements:** Go 1.22+, Node 20+, Docker (for the easy path), or local
MongoDB + Redis installs.

### Option A — Docker (recommended)

```bash
cp backend/.env.example backend/.env      # edit JWT_SECRET
docker compose up -d --build
```

- Frontend: http://localhost:5173
- Backend:  http://localhost:8080
- Health check: http://localhost:8080/api/health

### Option B — Run natively

```bash
# Terminal 1: infra
docker run -d -p 27017:27017 mongo:7
docker run -d -p 6379:6379 redis:7-alpine

# Terminal 2: backend
cd backend
cp .env.example .env
export GOPROXY=direct GOSUMDB=off   # only needed on networks that block the default Go module proxy
go mod tidy
go run ./cmd/server

# Terminal 3: frontend
cd frontend
cp .env.example .env
npm install
npm run dev
```

> **Note on `GOPROXY`:** this repo's `go.mod` includes `replace` directives
> that redirect several `golang.org/x/...`, `gopkg.in/...`, and
> `go.mongodb.org/...` vanity import paths to their GitHub mirrors. This was
> necessary to build in a network-restricted sandbox during development.
> On a normal machine with unrestricted internet access, you can delete
> the `replace (...)` block from `backend/go.mod` and run `go mod tidy` to
> resolve everything from the standard Go module proxy instead — either
> way works.

## Environment variables

**Backend** (`backend/.env`):

| Variable | Description |
|---|---|
| `PORT` | HTTP port (default `8080`) |
| `MONGODB_URI` | Mongo connection string |
| `MONGODB_DATABASE` | Database name |
| `REDIS_URL` | Redis connection string |
| `JWT_SECRET` | Secret used to sign JWTs — set a long random value in production |
| `CORS_ORIGIN` | Allowed frontend origin |

**Frontend** (`frontend/.env`):

| Variable | Description |
|---|---|
| `VITE_API_URL` | Base URL of the REST API |
| `VITE_WS_URL` | Base URL for WebSocket connections |

## API

**Auth**
```
POST /api/auth/register   { email, password, confirmPassword }
POST /api/auth/login      { email, password }
GET  /api/auth/me         (Bearer token)
```

**Polls**
```
POST   /api/polls              (auth) create a poll
GET    /api/polls/mine         (auth) list your polls
GET    /api/polls/:id                 public poll details
GET    /api/polls/:id/results         live vote tally
POST   /api/polls/:id/close    (auth, owner) close a poll
DELETE /api/polls/:id          (auth, owner) delete a poll
```

**Voting**
```
POST /api/polls/:id/vote   { optionId }   — works logged-in or anonymous
```

**Realtime**
```
GET /ws/polls/:id   — WebSocket, streams { type: "vote_update", ... }
                       and { type: "viewer_count", ... }
```

**Health**
```
GET /api/health   — checks Mongo + Redis connectivity
```

Responses are always `{ "success": true, "data": {...} }` or
`{ "success": false, "error": { "code", "message" } }`.

## Database design

- `users` — unique index on `email`
- `polls` — unique index on `publicId`, index on `creatorId`
- `votes` — index on `pollId`, **unique compound index** on
  `(pollId, voterIdentifier)` — this is the durable guarantee against
  duplicate votes; Redis `SETNX` is the fast pre-check in front of it.

`voterIdentifier` is a SHA-256 hash of either the authenticated user's id
or an anonymous visitor's IP+User-Agent fingerprint — never raw PII.

## Security

- Passwords hashed with bcrypt, never stored or returned in plaintext
- JWT auth (24h expiry) required for poll management endpoints
- All input validated server-side (never trusts client-side checks)
- Rate limiting on `/auth/*` and `/polls/:id/vote`
- CORS restricted to the configured frontend origin
- Basic security headers (`X-Content-Type-Options`, `X-Frame-Options`, etc.)
- No secrets committed — `.env` is gitignored, `.env.example` provided

## Error handling

The API returns structured errors (`POLL_NOT_FOUND`, `ALREADY_VOTED`,
`POLL_CLOSED`, `INVALID_OPTION`, `RATE_LIMITED`, etc.) and never leaks
internal error details or stack traces. If Redis is unreachable when
publishing an update, the vote still persists in MongoDB — only the
instant broadcast is missed, and the next poll load (or reconnect) picks
up the correct count from Mongo. The frontend shows a "Reconnecting…"
indicator when its WebSocket drops and retries automatically.

## Testing

```bash
cd backend
go test ./...
```

Included: password hashing, JWT issue/verify, and poll-creation
validation logic (question/option limits, duplicate detection). For a
full CI setup, add integration tests against a real MongoDB/Redis using
`testcontainers-go`, and add frontend tests (Vitest + React Testing
Library) for the login, poll-creation, and voting flows.

## Deployment

This repo is deployment-ready but has **not been deployed to a public
URL from this environment** — the sandbox this was built in has no
outbound network access beyond a small allowlist (npm/PyPI/GitHub), so
it can't reach a hosting provider, a managed MongoDB, or a managed
Redis instance. To deploy for real:

1. **MongoDB** — create a free cluster on MongoDB Atlas, get the
   connection string.
2. **Redis** — create an instance on Upstash, Redis Cloud, or similar.
3. **Backend** — deploy `backend/` (with its `Dockerfile`) to Render,
   Railway, Fly.io, or similar. Set `MONGODB_URI`, `REDIS_URL`,
   `JWT_SECRET`, `CORS_ORIGIN` as environment variables.
4. **Frontend** — deploy `frontend/` to Vercel, Netlify, or the same
   Docker host. Set `VITE_API_URL` / `VITE_WS_URL` to your backend's
   public HTTPS/WSS URL at build time.
5. Verify with the multi-browser test below.

### Production realtime verification

Open the same poll URL in two or three browser windows/devices. Vote in
one. The others must update immediately, with no refresh. Then test:
multiple votes, multiple options, a closed poll, an invalid vote, a
network drop/reconnect, and a mobile browser.

## AI usage

This project's code, structure, and this README were generated with AI
assistance (Claude) working from a detailed assignment specification,
in an interactive session that also compiled and test-ran the Go
backend and built the React frontend to verify correctness before
delivery. Architectural decisions (validation rules, dedupe strategy,
Redis/WebSocket bridging, error taxonomy) were made by working through
the spec directly rather than accepting unreviewed output — if you're
presenting this, make sure you can explain each of the "Why" points
below in your own words.

**Why these choices:**
- **MongoDB as source of truth, Redis as cache:** vote counts must
  survive a Redis restart; the unique compound index in Mongo is the
  real duplicate-vote guarantee, Redis `SETNX` is just a fast pre-check.
- **One Redis subscription per poll, not per client:** avoids Redis
  connection/subscription exhaustion under many concurrent viewers of
  the same poll.
- **Hashed voter identifiers:** lets us block duplicate votes without
  storing raw IP addresses or personal data on the vote record.
- **Alternatives considered:** polling every N seconds instead of
  WebSockets (rejected — spec explicitly forbids this and it doesn't
  scale); storing counts only in Mongo with no cache (rejected — every
  page load would hit an aggregation query instead of an O(1) hash read).
