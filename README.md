# clean-arc-eg

A minimal Go example of **Clean Architecture** for a simple Budget service backed by MongoDB.
It wires **repository → usecase → HTTP** with context cancellation/timeout handling and graceful shutdown.

> Repo layout (as of latest): `src/` package plus `main.go`, `go.mod`, `go.sum`. :contentReference[oaicite:0]{index=0}

---

## Features

- Clean separation:
  - **Repository**: MongoDB driver (read all budgets)
  - **Usecase**: business logic + context error mapping
  - **Transport**: HTTP handler for `GET /budgets`
- Context-aware error handling (`context.Canceled`, `context.DeadlineExceeded`)
- Sensible HTTP codes (405, 499/408, 504, 500)
- JSON responses with `Cache-Control: no-store` and server timestamp
- Graceful shutdown

---

## Architecture

```

client ──HTTP──> handler (src/http.go)
│
▼
usecase (src/usecase.go)
│   BudgetRepository interface
▼
repository (src/mongo.go) ──> MongoDB

```

Key contracts:

- `type BudgetRepository interface { GetAllBudget(ctx context.Context) ([]Budget, error) }`
- `type Usecase struct { repo BudgetRepository }`

---

## Project layout

```

.
├─ main.go
├─ go.mod
├─ go.sum
└─ src/
├─ http.go        # HTTP transport (GET /budgets)
├─ mongo.go       # Mongo repository implementation
└─ usecase.go     # Usecase and domain models

````

---

## Getting started

### Prerequisites
- Go 1.22+ (recommended)
- MongoDB 6.x+ running locally or reachable via URI

### Configuration (env)

| Var                | Default                     | Description                  |
|--------------------|-----------------------------|------------------------------|
| `MONGO_URI`        | `mongodb://localhost:27017` | MongoDB connection string    |
| `MONGO_DB`         | `budgetdb`                  | Database name                |
| `MONGO_COLLECTION` | `budgets`                   | Collection name              |
| `ADDR`             | `:8080`                     | HTTP listen address          |

> You can create a `.env` file and the app will load it if present.

### Run

```bash
export MONGO_URI="mongodb://localhost:27017"
export MONGO_DB="budgetdb"
export MONGO_COLLECTION="budgets"
export ADDR=":8080"

go run ./...
````

You should see:

```
HTTP server listening on :8080
```

Stop with `Ctrl+C` (graceful shutdown).

---

## API

### GET `/budgets`

Fetch all budgets.

#### Response `200 OK`

```json
{
  "budgets": [
    {
      "id": "66ef...e0",
      "name": "Ops",
      "amount": 12000,
      "currency": "USD",
      "lastUpdated": "2025-09-24T04:20:31.123Z"
    }
  ],
  "timestamp": "2025-09-24T04:20:31.456Z"
}
```

#### Error responses

* `405 Method Not Allowed` — non-GET methods
* `499 Client Closed Request` (or `408`) — client cancelled request
* `504 Gateway Timeout` — upstream timeout
* `500 Internal Server Error` — other failures

#### cURL

```bash
curl -i http://localhost:8080/budgets
```

---

## Data model

Domain (usecase):

```go
type Budget struct {
  ID          string
  Name        string
  Amount      float64
  Currency    string
  LastUpdated time.Time
}
```

Mongo document (`src/mongo.go`) maps `_id` to `ID` (hex).

---

## Development notes

* **Context**: both HTTP layer and repository respect `r.Context()`. Use timeouts if needed.
* **HTTP codes**: handler translates well-known context errors; everything else bubbles as 500.
* **Caching**: responses set `Cache-Control: no-store` to prevent stale API caching.
* **Extending**:

    * Add new use cases by growing `Usecase` and repository interfaces.
    * Add routes in `main.go` mux.

---

## Seeding data (mongo shell)

```js
use budgetdb
db.budgets.insertMany([
  { name: "Ops", amount: 12000, currency: "USD", last_updated: new Date() },
  { name: "R&D", amount: 30000, currency: "USD", last_updated: new Date() }
])
```

---

## License

MIT (or your preferred license). Add a `LICENSE` file if you want open-source distribution.

---

## Attributions

* Repo: `pamungkaski/clean-arc-eg` (structure and filenames). ([GitHub][1])
