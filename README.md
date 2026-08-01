# Distributed Real-Time Multiplayer Tic-Tac-Toe (Cloudfactory Backend coding challenge)

A distributed, real-time multiplayer Tic-Tac-Toe platform built around a Go backend and an Electron/React desktop client. The server handles matchmaking, session management, the game engine, persistence, and observability; every running instance of the desktop client represents one connected participant.

This repository contains two independently deployable applications - `server/` and `client/`, plus the infrastructure required to run the full stack locally with Docker Compose.

---

## Table of Contents

- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Features](#features)
- [High-Level Design](#high-level-design)
  - [Desktop Client](#desktop-client)
  - [Backend server](#go-server)
  - [PostgreSQL](#postgresql)
  - [Redis / In-Memory Map](#redis--in-memory-map)
  - [Flyway](#flyway)
  - [Prometheus](#prometheus)
  - [Grafana](#grafana)
- [Domain Model](#domain-model)
- [Entity Relationship Diagram](#entity-relationship-diagram)
- [Sequence Diagrams](#sequence-diagrams)
  - [1. Startup, Connection &amp; Lobby](#1-startup-connection--lobby)
  - [2. Creating a Game Session (Invitation)](#2-creating-a-game-session-invitation)
  - [3. In-Game Play](#3-in-game-play)
  - [4. Resuming Game Play (Reconnection &amp; Session Recovery)](#4-resuming-game-play-reconnection--session-recovery)
- [Project Structure](#project-structure)
- [Configuration](#configuration)
  - [Client Configuration](#client-configuration)
  - [Server Configuration](#server-configuration)
- [Running the Project](#running-the-project)
  - [Prerequisites](#prerequisites)
  - [Start Infrastructure](#start-infrastructure)
  - [Run Client](#run-client)
- [Development Workflow](#development-workflow)
- [Monitoring](#monitoring)
- [Logging](#logging)
- [Error Handling](#error-handling)
- [Testing](#testing)
- [Production Recommendations](#production-recommendations)
  - [Logging](#logging-1)
  - [Database](#database)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

---

## Architecture

```text
                                +-------------------------+
                                |     Electron Client     |
                                |   React + TypeScript    |
                                +-------------------------+
                                             |
                                         WebSocket
                                             v
  +-------------------------------------------------------------------------------------+
  |                                  Go Backend Server                                  |
  +-------------------------------------------------------------------------------------+
  | WebSocket Gateway | Matchmaking | Session Manager | Game Engine | Metrics | Logging |
  +-------------------------------------------------------------------------------------+
                |                            |                            |
                v                            v                            v
        +--------------+            +---------------+             +--------------+
        |  PostgreSQL  |            |  Redis / Map  |             |  Prometheus  |
        +--------------+            +---------------+             +--------------+
                ^                                                         |
                |                                                         v
        +--------------+                                           +------------+
        |    Flyway    |                                           |  Grafana   |
        |DB Migrations |                                           +------------+
        +--------------+
```

**Component responsibilities**

| Component                       | Responsibility                                                                                                                                                                                                                                                                                               |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Electron Client**       | Renders the lobby and game board, opens and maintains a single WebSocket connection to the server, and drives reconnection/session-recovery logic on the client side.                                                                                                                                        |
| **WebSocket Gateway**     | Terminates client WebSocket connections, authenticates/registers participants, and multiplexes inbound/outbound messages to the appropriate internal subsystem.                                                                                                                                              |
| **Matchmaking**           | Pairs two participants into a session via a direct invitation from one to the other. There is no automatic matchmaking queue; a participant can only start a game by inviting (or being invited by) someone.                                                                                               |
| **Session Manager**       | Owns the lifecycle of an active game session, creation, heartbeat tracking, pause on disconnect, and recovery on reconnect.                                                                                                                                                                                |
| **Game Engine**           | Validates and applies moves, enforces turn order, and detects win/draw conditions for each session.                                                                                                                                                                                                          |
| **PostgreSQL**            | Durable system of record for participants, completed sessions, move history, and invitations.                                                                                                                                                                                                                |
| **Redis / In-Memory Map** | Transient state, active connections, presence, invitations, and heartbeats. This demo runs as a single Go process, so it keeps this state in a plain in-memory map; Redis is a drop-in replacement implementing the same interface when the server is horizontally scaled. Never used for durable storage. |
| **Flyway**                | Applies versioned schema migrations to PostgreSQL automatically on container startup, before the Backend server begins accepting traffic.                                                                                                                                                                    |
| **Prometheus**            | Scrapes the Backend server's`/metrics` endpoint on a fixed interval and stores time-series metrics.                                                                                                                                                                                                        |
| **Grafana**               | Visualizes metrics collected by Prometheus across connection, game, and infrastructure dashboards.                                                                                                                                                                                                           |

---

## Technology Stack

| Technology                                               | Purpose                                                                                                                                   |
| -------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| Go                                                       | Backend server runtime and language                                                                                                       |
| [WebSocket protocol](https://github.com/nhooyr/websocket) | WebSocket protocol implementation                                                                                                         |
| PostgreSQL                                               | Durable persistent storage                                                                                                                |
| Redis / in-memory map                                    | Active sessions, presence, and invitations; an in-memory map in this single-instance demo, swappable for Redis when horizontally scaled |
| Flyway                                                   | Database schema migrations                                                                                                                |
| Prometheus                                               | Metrics collection                                                                                                                        |
| Grafana                                                  | Metrics dashboards and visualization                                                                                                      |
| Docker                                                   | Containerization of all services                                                                                                          |
| Docker Compose                                           | Local multi-service orchestration                                                                                                         |
| Electron                                                 | Desktop application runtime                                                                                                               |

---

## Features

- Real-time multiplayer Tic-Tac-Toe over WebSockets
- Invitation-only matchmaking; a participant can only start a game by inviting (or being invited by) another participant; there is no automatic matchmaking queue
- Lobby showing online participants and pending invitations
- Automatic reconnection with exponential backoff and random jitter
- Session recovery after a dropped connection resumes the in-progress game
- Heartbeat monitoring to detect stale or dead connections
- Versioned, auto-applied database migrations
- Full observability stack: metrics, dashboards, and centralized structured logging
- Fully containerized local deployment via Docker Compose

---

## High-Level Design

### Desktop Client

**Responsibilities**

- Load connection and reconnection configuration at startup
- Open and maintain a single WebSocket connection to the Backend server
- Render the lobby (online participants and pending invitations)
- Render the active game board and reflect moves in real time
- Detect disconnects and drive reconnection with exponential backoff + jitter
- Request and apply session recovery data after a successful reconnect

### Backend server

**Responsibilities**

- Accept and upgrade incoming WebSocket connections
- Register newly connected participants and track their presence in the in-memory map (Redis if horizontally scaled)
- Maintain the set of active connections and detect disconnects via heartbeats
- Process direct game invitations, the only way a game session can start
- Process, validate, and apply game moves through the game engine
- Detect win/draw conditions and finalize completed sessions
- Persist completed games and moves to PostgreSQL
- Expose a `/metrics` endpoint for Prometheus scraping
- Emit structured JSON logs for every significant lifecycle event

### PostgreSQL

Durable system of record. Stores:

- Participants
- Sessions
- Moves
- Invitations

### Redis / In-Memory Map

This demo runs as a **single Backend server instance**, so there is no cross-instance state to synchronize. The transient state below lives in a plain in-memory map inside that one process. Redis is the pluggable, production-grade alternative implementing the same interface; swap it in when the server is horizontally scaled across multiple instances and that state needs to be shared. Either way, this store is transient only **not** used for durable storage. Stores:

- Active WebSocket connections
- Online participant presence
- Pending invitations
- Active session pointers
- Heartbeat timestamps

### Flyway

Flyway manages all PostgreSQL schema migrations. It runs as a short-lived container that executes automatically on stack startup, applying any pending migrations before the Backend server container is allowed to begin accepting traffic (enforced via a Docker Compose `depends_on` health/completion condition).

```text
migrations/
├── V1__initial_schema.sql
├── V2__create_sessions.sql
├── V3__create_moves.sql
├── V4__leaderboard.sql
└── V5__drop_leaderboard.sql
```

Migrations are an append-only log, once applied, a file is never edited or deleted, since Flyway validates already-applied migrations by checksum. `V4` is a good example: the leaderboard feature it introduced was later removed, but rather than rewriting or deleting `V4`, `V5` was added to drop what it created.

### Prometheus

The Backend server exposes a `/metrics` endpoint in Prometheus exposition format. Prometheus is configured to scrape this endpoint on a fixed interval and stores the resulting time series for querying and alerting.

### Grafana

Grafana visualizes the metrics collected by Prometheus through pre-provisioned dashboards covering:

- Connections
- Games
- Latency
- Infrastructure

---

## Domain Model

| Entity                | Responsibility                                                                                                                              |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| **Participant** | A registered player identity; the anchor for presence, sessions, and invitations.                                                           |
| **Connection**  | A single live WebSocket connection belonging to a participant; tracked transiently in the in-memory map (or Redis, if horizontally scaled). |
| **Session**     | An active or completed game between two participants, including turn state and outcome.                                                     |
| **Move**        | A single placed mark within a session, recorded in order for replay and history.                                                            |
| **Invitation**  | A pending or resolved direct challenge from one participant to another.                                                                     |

---

## Entity Relationship Diagram

```text
  +-------------+      +--------------------------+      +-----------------------+
  | Participant |      | Invitation               |      | Session               |
  +-------------+      +--------------------------+      +-----------------------+
  | id (PK)     |      | id (PK)                  |      | id (PK)               |
  | username    |      | from_participant_id (FK) |      | participant_x_id (FK) |
  | created_at  |      | to_participant_id (FK)   |      | participant_o_id (FK) |
  +-------------+      | status                   |      | status                |
                       | created_at               |      | winner_id (FK)        |
                       +--------------------------+      | started_at            |
                                                          | completed_at          |
                                                          +-----------------------+
         |                           |                               |
     1 initiates N              1 accepted-as 1
         +---------------------------+-------------------------------+

                                                                     | 1
                                                                     | has
                                                                     v N
                                                          +---------------------+
                                                          | Move                |
                                                          +---------------------+
                                                          | id (PK)             |
                                                          | session_id (FK)     |
                                                          | participant_id (FK) |
                                                          | position            |
                                                          | move_number         |
                                                          | created_at          |
                                                          +---------------------+
```

---

## Sequence Diagrams

### 1. Startup, Connection & Lobby

```text
Client                      Server                      PostgreS[ User enters name, clicks Connect ]
|                           |                           |
|---- hello (username) ---->|                           |
|                           | -- Register participant --|
|                           |--- Upsert Participant --->|
|< welcome (participantId) -|                           |
|<------ lobby_state -------|                           |
                            [ Broadcast presence_update(online) to others ]
```

1. The Electron client shows the "New Connection" dialog on launch and does not open a WebSocket connection until the user submits a participant name.
2. The client sends `hello` with that name (and its persisted `participantId`, if this isn't the first launch).
3. The server registers the participant in the in-memory store, records a heartbeat, and upserts the participant row in PostgreSQL.
4. The server replies with `welcome` (confirming the participant ID to persist locally) and `lobby_state` (currently online, invitable participants and pending invitations).
5. The server broadcasts `presence_update` so every other connected client sees this participant appear in their lobby.

---

### 2. Creating a Game Session (Invitation)

There is no automatic matchmaking queue; a game can only start when one participant invites another and that invitation is accepted.

#### Invitation accepted

```text
Inviter                    Server                     Invitee                    PostgreSQL
|                          |                          |                          |
|----- invite_create ----->|                          |                          |
|                          | -- Save invitation --    |                          |
|<----- invite_sent -------|                          |                          |
|                          |---- invite_received ---->|                          |
|                          |                          |                          |
|                          |<-------------------------|                          |
|                          | -- Delete invitation --  |                          |
|                          |------------------ Insert Session ------------------>|
|<----- game_started ------|                          |                          |
|                          |------ game_started ----->|                          |
                           [ Broadcast presence_update(unavailable) for both ]
```

1. The inviter picks a participant from the "Play Game" dialog; the server saves a pending invitation and confirms it to the inviter (`invite_sent`) while notifying the invitee (`invite_received`).
2. The invitee accepts (`invite_respond`, `accept: true`). The server deletes the pending invitation, creates the session, and persists it to PostgreSQL.
3. Both participants receive `game_started` with the new session (empty board, inviter plays `X`).
4. The server broadcasts that both participants are no longer available for invitations, so they drop out of everyone else's "Play Game" list until the game ends.

#### Invitation declined

```text
Inviter                     Server                      Invitee
|                           |                           |
|------ invite_create ----->|                           |
|                           | -- Save invitation --     |
|<------ invite_sent -------|                           |
|                           |----- invite_received ---->|
|                           |                           |
|                           |<--------------------------|
|                           | -- Delete invitation --   |
|<---- invite_canceled -----|                           |
```

1. Same start as above, but the invitee responds with `invite_respond`, `accept: false`.
2. The server deletes the invitation and notifies only the inviter with `invite_canceled`, no session is created, and neither participant's availability changes.

---

### 3. In-Game Play

```text
Player A                  Server                    Player B                  PostgreSQL
|                         |                         |                         |
|---- move (position) --->|                         |                         |
|                         | -- Validate turn + apply --                       |
|                         |------------------- Insert Move ------------------>|
|<---- move_applied ------|                         |                         |
|                         |----- move_applied ----->|                         |
|                         |                         |                         |
                          [ ... repeated for each move ... ]
|                         |                         |                         |
|                         |<------------------------|                         |
|                         | -- Detect win --        |                         |
|                         |---------------- Complete Session ---------------->|
|<--- game_completed -----|                         |                         |
|                         |---- game_completed ---->|                         |
                          [ Broadcast presence_update(available) for both ]
```

1. A player sends `move` with the session ID and board position. The server rejects it (with an `error` message, not shown here) unless the sender is a participant in that session and it's currently their turn.
2. A valid move is applied to the in-memory board, persisted as a `Move` row, and broadcast to both players as `move_applied`, this repeats for every move in the game.
3. When a move completes three in a row (or fills the board with no winner), the server marks the session completed and persists the final result.
4. Both players receive `game_completed`, and the server broadcasts that they're available for invitations again.

---

### 4. Resuming Game Play (Reconnection & Session Recovery)

```text
Client                                      Server
|                                           |
X===========================================X  Connection Lost
|                                           | -- Unregister + mark offline --
                                            [ Active session left untouched in store ]
|                                           |
[ Exponential backoff + jitter ]
|------- hello (same participantId) ------->|
|                                           | -- Recover active session --
|<--------- welcome (reconnected) ----------|
[ Client renders GameBoard directly, skips lobby ]
```

1. The connection drops mid-game (network blip, app backgrounded, etc.). The server notices the closed socket, unregisters the connection, and marks the participant offline; but deliberately leaves their active session exactly as it was in the store, so it can be recovered.
2. The client's `WSClient` retries with exponential backoff and jitter (capped by `maxReconnectDelayMs`, up to `maxReconnectAttempts`), sending the same persisted `participantId` it always uses.
3. The server recognizes the returning participant, looks up their active session, and returns it in `welcome.activeSession`.
4. The client skips the lobby entirely and renders the game board directly from that recovered session, board, turn, and both players' identities all restored.

---

## Project Structure

```text
project/
├── client/           # Electron + React + TypeScript desktop application
├── server/           # Go backend: gateway, matchmaking, game engine, persistence
├── docker/           # Dockerfiles for the server and supporting images
├── migrations/        # Flyway-managed SQL migration scripts
├── monitoring/
│   ├── grafana/        # Provisioned dashboards and datasources
│   └── prometheus/     # Scrape configuration
├── docker-compose.yml  # Orchestrates the full local stack
└── README.md
```

| Directory                  | Contents                                                                           |
| -------------------------- | ---------------------------------------------------------------------------------- |
| `client/`                | React UI, Zustand stores, WebSocket client, Electron main/preload processes        |
| `server/`                | WebSocket gateway, matchmaking, session manager, game engine, persistence, metrics |
| `docker/`                | Multi-stage Dockerfiles for building the server and client images                  |
| `migrations/`            | Ordered Flyway SQL migrations defining the PostgreSQL schema                       |
| `monitoring/grafana/`    | Dashboard JSON definitions and the Prometheus datasource config                    |
| `monitoring/prometheus/` | `prometheus.yml` scrape target configuration                                     |

---

## Configuration

### Client Configuration

```json
{
  "serverUrl": "ws://localhost:8080/ws",
  "maxReconnectAttempts": 20,
  "baseReconnectDelayMs": 500,
  "maxReconnectDelayMs": 10000
}
```

### Server Configuration

| Variable              | Purpose                                                                                                          |
| --------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `POSTGRES_HOST`     | PostgreSQL host                                                                                                  |
| `POSTGRES_PORT`     | PostgreSQL port                                                                                                  |
| `POSTGRES_USER`     | PostgreSQL username                                                                                              |
| `POSTGRES_PASSWORD` | PostgreSQL password                                                                                              |
| `POSTGRES_DATABASE` | PostgreSQL database name                                                                                         |
| `REDIS_HOST`        | Redis host (optional, only used when running with the Redis-backed store instead of the default in-memory map) |
| `REDIS_PORT`        | Redis port (optional, see`REDIS_HOST`)                                                                       |
| `LOG_LEVEL`         | Minimum log level (`debug`, `info`, `warn`, `error`)                                                     |
| `SERVER_PORT`       | Port the Backend server listens on                                                                               |

---

## Running the Project

### Prerequisites

- Node.js v22.14.0
- npm or Yarn
- Docker (Docker Desktop, which includes Docker Compose)
- Go SDK 1.24 or later

### Start Infrastructure

```bash
docker compose up -d
```

This starts the full backing stack:

- PostgreSQL
- Flyway (runs migrations, then exits)
- Backend server
- Prometheus
- Grafana

This demo runs a single Backend server instance, so it keeps active connections, presence, and pending invitations in an in-memory map, no Redis container is required. A `redis` service can be added to `docker-compose.yml` and pointed to via `REDIS_HOST`/`REDIS_PORT` if the server is horizontally scaled and that state needs to be shared across instances.

### Run Client

```bash
npm install
npm run dist:win
```

I recommend using Yarn for npm builds:

```bash
Yarn install
yarn run dist:win
```

---

## Development Workflow

1. Start the infrastructure with `docker compose up -d`.
2. Wait for the Flyway migration container to complete successfully.
3. Verify the Backend server is healthy (check `/metrics` or container logs).
4. Launch the Electron client with `npm run electron:dev`.
5. Open multiple client instances to simulate multiple participants.
6. Play games between instances to exercise invitations, moves, and completion.
7. Observe live connection, game, and infrastructure metrics in Grafana.

---

## Monitoring

### Connection Dashboard

- Active clients
- Connected users
- Reconnection rate
- Connection failures

### Game Dashboard

- Active games
- Completed games
- Average game duration

### Performance Dashboard

- CPU usage
- Memory usage
- Goroutine count
- Database query latency
- State store latency (in-memory map / Redis)
- WebSocket message latency

---

## Logging

The backend server emits structured JSON logs to stdout for all significant events, picked up by Docker's local log driver; view them with `docker compose logs -f server`. There's no log aggregation or centralized search in this local setup; see [Production Recommendations](#production-recommendations) for how to add that.

Example fields:

- `timestamp`
- `request_id`
- `participant_id`
- `session_id`
- `event_type`
- `latency`
- `message`

---

## Error Handling

| Scenario             | Behavior                                                                                                       |
| -------------------- | -------------------------------------------------------------------------------------------------------------- |
| Invalid message      | Rejected with a structured error response; connection remains open.                                            |
| Lost connection      | Marked disconnected after a missed heartbeat; associated session is paused, not terminated.                    |
| Retry exhaustion     | Client stops reconnecting after`maxReconnectAttempts` and surfaces a persistent connection-failure state.    |
| Invalid move         | Rejected by the game engine with an error message; game state is unchanged.                                    |
| Server unavailable   | Client enters the reconnection loop with exponential backoff and jitter.                                       |
| Database unavailable | Server logs the failure, degrades game-completion persistence, and reports the issue via`/metrics` and logs. |

---

## Testing

Run all server tests:

```bash
go test ./...
```

Run client tests:

```bash
npm test
```

Check server test coverage:

```bash
go test -cover ./...
```

Lint:

```bash
golangci-lint run
npm run lint
```

---

## Production Recommendations

This repository is set up as a local demo. A few things are worth calling out explicitly before running any of this in production.

### Logging

The server currently just writes structured JSON logs to stdout, captured by Docker's local log driver; fine for `docker compose logs`, but there's no aggregation, retention, or cross-container search. **Recommendation: add a log aggregation system**, such as [Grafana Loki](https://grafana.com/oss/loki/) (paired with Promtail or the [Docker driver plugin](https://grafana.com/docs/loki/latest/send-data/docker-driver/)) shipping into the same Grafana instance already used for metrics, or a managed provider (CloudWatch Logs, Datadog, etc.). This project ran with a local Loki + Promtail setup earlier in development; it was removed to keep the default stack smaller, but the JSON log format (`timestamp`, `request_id`, `participant_id`, `session_id`, `event_type`, `latency`, `message`) is already structured for exactly this.

### Database

1. **Use connection pooling.** ✅ Already in place `internal/db/postgres.go` connects via [`pgxpool`](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool) rather than a single `*sql.DB` connection. For production, don't just rely on pgxpool's defaults: size the pool explicitly (`pool_max_conns` / `pool_min_conns` in the DSN, or via `pgxpool.ParseConfig`) to match expected concurrency and stay under PostgreSQL's own `max_connections`.
2. **Index `session_id` on the `moves` table.** ✅ Already in place `idx_moves_session` in `migrations/V3__create_moves.sql`, since every move lookup and session replay filters by `session_id`.
