# key value store

A lightweight, persistent key-value store written in Go with an HTTP API. Data is durably stored to disk using an append-only file format with automatic background compaction and garbage collection.

## Features

- **HTTP REST API** — Simple JSON endpoints for `SET`, `GET`, and `DELETE` operations
- **Persistent storage** — All writes are appended to disk and survive server restarts
- **Automatic restore** — The in-memory index is rebuilt from the data file on startup
- **Background compaction** — A periodic goroutine rewrites the data file to remove stale entries
- **Background deletion** — Deletes are batched and asynchronously purged from the data file
- **Thread-safe** — All operations are protected by mutexes for safe concurrent access
- **Zero dependencies** — Built entirely on the Go standard library

## Architecture

```
┌──────────────────────────────────────────────────┐
│                   HTTP Server (:8080)             │
│         /set (POST)  /get (GET)  /delete (DELETE) │
└──────────────────┬───────────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────────┐
│                    Engine                         │
│                                                   │
│  ┌──────────────┐   ┌────────────┐               │
│  │ In-Memory    │   │ Append-Only│               │
│  │ Index (map)  │◄──│ Data File  │               │
│  │ key → offset │   │ (data.txt) │               │
│  └──────────────┘   └────────────┘               │
│                                                   │
│  Background Goroutines:                           │
│    • CompactFile()  — rewrites data.txt every 5s  │
│    • DeleteFromFile() — purges deleted keys every 5s│
│                                                   │
│  ┌──────────────┐                                 │
│  │ Delete Queue  │                                │
│  │ (delete.txt)  │                                │
│  └──────────────┘                                 │
└──────────────────────────────────────────────────┘
```

### How it works

1. **Set** — The key-value pair is appended to `data.txt` as `key value\n`. The byte offset is stored in an in-memory map for O(1) lookups.
2. **Get** — The engine looks up the byte offset for the key, seeks to that position in the file, and reads the value.
3. **Delete** — The key is removed from the in-memory index immediately and appended to `delete.txt`. A background goroutine periodically reads `delete.txt`, removes those keys from `data.txt`, and truncates the delete file.
4. **Compaction** — Every 5 seconds, the data file is scanned and rewritten with only the latest value per key, reclaiming space from overwrites.
5. **Restore** — On startup, the entire data file is scanned to rebuild the in-memory offset index so previously stored data is available immediately.

### Storage location

Data files are stored in `~/.config/keyvaluedb/`:

```
~/.config/keyvaluedb/
├── data.txt       # Append-only key-value data
└── delete.txt     # Pending delete queue
```

## Getting Started

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)

### Build & Run

```bash
# Clone the repository
git clone https://github.com/vivek1504/key-value-go.git
cd key-value-go

# Run directly
go run .

# Or build a binary
go build -o keyvaluedb .
./keyvaluedb
```

The server starts on **http://localhost:8080**.

## API Reference

### Set a key-value pair

```
POST /set
Content-Type: application/json
```

**Request body:**

```json
{
  "key": "username",
  "value": "alice"
}
```

**Success response** (`200 OK`):

```json
{
  "key": "success",
  "value": "Key value pair saved successfully."
}
```

**Error response** (`500 Internal Server Error`):

```json
{
  "key": "error",
  "value": "key cannot contain spaces"
}
```

> **Note:** Keys cannot contain spaces.

### Get a value by key

```
GET /get?key=username
```

**Success response** (`200 OK`):

```json
{
  "key": "username",
  "value": "alice"
}
```

**Error response** (`404 Not Found`):

```json
{
  "key": "error",
  "value": "key not found"
}
```

### Delete a key

```
DELETE /delete?key=username
```

**Success response** (`200 OK`):

```json
{
  "key": "success",
  "value": "Key deleted successfully."
}
```

### Quick examples with cURL

```bash
# Set a value
curl -X POST http://localhost:8080/set \
  -H "Content-Type: application/json" \
  -d '{"key": "language", "value": "go"}'

# Get a value
curl http://localhost:8080/get?key=language

# Delete a value
curl -X DELETE http://localhost:8080/delete?key=language
```

## Running Tests

```bash
go test -v ./...
```

The test suite covers:

| Test | Description |
|------|-------------|
| `Test_SetGetKeyValue` | Basic set and get operations, including missing key errors |
| `TestEngine_Compact` | Verifies compaction deduplicates overwritten keys |
| `TestEngine_Restore` | Ensures the index is rebuilt correctly after restart |
| `TestEngine_DeleteKey` | Validates in-memory deletion and delete queue writes |
| `TestEngine_DeleteKeyFromFile` | Confirms physical removal of keys from the data file |

## Project Structure

```
.
├── main.go            # HTTP server, route handlers, and entry point
├── engine.go          # Storage engine: persistence, compaction, deletion
├── engine_test.go     # Test suite
├── go.mod             # Go module definition
├── .gitignore         # Ignores data.txt and delete.txt
└── readme.md          # This file
```

## Design Decisions

- **Append-only writes** — Writes never modify existing data in-place, reducing the risk of corruption and simplifying the write path.
- **Offset-based reads** — Values are read directly from disk via `seek` rather than being cached in memory, keeping the memory footprint small (only keys and offsets are held in RAM).
- **Separate delete queue** — Deletes are immediately reflected in the in-memory index but physically removed from the data file asynchronously, keeping delete latency low.
- **Periodic compaction** — Rather than compacting on every write, a background loop runs every 5 seconds to balance write throughput with disk space reclamation.
