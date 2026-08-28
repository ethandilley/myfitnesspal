# myfitnesspal CLI

A command-line client for the `myfitnesspal` gRPC service — track foods and log daily macros.

The CLI is a thin wrapper: every command opens a gRPC connection to the server and calls one RPC. It does no local storage of its own, so the server (and its Postgres database) must be running for any command to work.

---

## 1. Setup

### 1.1 Start the server + database

The easiest path is Docker Compose, which builds the server image and spins up Postgres:

```bash
make compose
```

This runs `docker compose up --build -d`, which starts:
- `server` — the gRPC service, exposed on `localhost:50051`
- `postgres` — Postgres 18, exposed on `localhost:5432` (db `myfitnesspal`, user/pass `user`/`password`)

To run the server directly instead (e.g. for local development against a Postgres instance you manage yourself):

```bash
export DB_URL="postgres://user:password@localhost:5432/myfitnesspal?sslmode=disable"
make server
# equivalent to: go run cmd/server/main.go
```

The server reads `DB_URL` from the environment and will exit immediately if it's not set.

### 1.2 Run database migrations

Schema migrations live in `db/migrations` and are applied with [goose](https://github.com/pressly/goose):

```bash
make migrate DB_URL="postgres://user:password@localhost:5432/myfitnesspal?sslmode=disable"
```

### 1.3 Run the CLI

```bash
make cli -- <command> [flags]
# equivalent to: go run cmd/cli/main.go <command> [flags]
```

Or build/install it as a binary if you'd rather call it directly as `myfitnesspal`:

```bash
go build -o myfitnesspal ./cmd/cli
./myfitnesspal <command> [flags]
```

---

## 2. Global flags

Every command inherits this persistent flag from the root command:

| Flag | Default | Description |
|---|---|---|
| `--addr` | `localhost:50051` | Address of the gRPC server to connect to |

Use it to point the CLI at a non-default server, e.g. a remote host or a different port:

```bash
myfitnesspal food ls --addr myserver.example.com:50051
```

---

## 3. Command tree

```
myfitnesspal
├── food
│   ├── add     Create a new food
│   ├── ls      List all foods
│   └── rm      Delete a food by id
└── log
    ├── add     Log a food
    ├── rm      Delete a log entry by id
    ├── ls      List all log entries
    └── today   Show a day's log entries and macro totals
```

---

## 4. `food` — manage foods

Foods are the reusable items you log against (e.g. "Chicken Breast", "Oatmeal"). Each food stores calories and macros **per serving**; the multiplier applied at log time scales those values.

### 4.1 `food add`

Create a new food.

```bash
myfitnesspal food add --name "Chicken Breast" --calories 165 --protein 31 --carbs 0 --fat 3.6
```

| Flag | Type | Required | Default | Description |
|---|---|---|---|---|
| `--name` | string | **yes** | — | Food name |
| `--calories` | float | no | `0` | Calories per serving |
| `--protein` | float | no | `0` | Protein in grams per serving |
| `--carbs` | float | no | `0` | Carbs in grams per serving |
| `--fat` | float | no | `0` | Fat in grams per serving |

Output:

```
created food #1: Chicken Breast
```

Note the returned id — you'll need it for `log add`.

### 4.2 `food ls`

List every food in the database.

```bash
myfitnesspal food ls
```

Output:

```
#1  Chicken Breast        cal:165  protein:31.0g  carbs:0.0g  fat:3.6g
#2  Oatmeal                cal:150  protein:5.0g  carbs:27.0g  fat:2.5g
```

No flags.

### 4.3 `food rm`

Delete a food by its numeric id.

```bash
myfitnesspal food rm 2
```

| Argument | Required | Description |
|---|---|---|
| `id` (positional) | **yes** — exactly one | The food's numeric id |

Output:

```
deleted food #2
```

---

## 5. `log` — log foods and view daily macros

Log entries link a food to a date, with a multiplier for serving size (e.g. `2` for a double portion, `0.5` for half).

### 5.1 `log add`

Log a food entry.

```bash
myfitnesspal log add --food-id 1 --multiplier 2 --date 2026-08-27
```

| Flag | Type | Required | Default | Description |
|---|---|---|---|---|
| `--food-id` | int32 | **yes** | `0` | Id of the food being logged (from `food ls`) |
| `--multiplier` | float | no | `1` | Serving multiplier applied to the food's macros |
| `--date` | string | no | *(empty → today)* | Date in `YYYY-MM-DD` format |

Output:

```
logged entry #7
```

### 5.2 `log ls`

List every log entry ever recorded, across all dates.

```bash
myfitnesspal log ls
```

Output:

```
#7  food:1  x2.0  2026-08-27
#8  food:2  x1.0  2026-08-27
```

No flags.

### 5.3 `log rm`

Delete a log entry by its numeric id.

```bash
myfitnesspal log rm 7
```

| Argument | Required | Description |
|---|---|---|
| `id` (positional) | **yes** — exactly one | The log entry's numeric id |

Output:

```
deleted log entry #7
```

### 5.4 `log today`

Show log entries for a single day plus the totaled macros for that day.

```bash
# today
myfitnesspal log today

# a specific past/future date
myfitnesspal log today --date 2026-08-27
```

| Flag | Type | Required | Default | Description |
|---|---|---|---|---|
| `--date` | string | no | *(empty → today)* | Date in `YYYY-MM-DD` format |

Output:

```
#7  food:1  x2.0
#8  food:2  x1.0

totals: cal:480  protein:67.0g  carbs:27.0g  fat:9.7g
```

---

## 6. Typical workflow

```bash
# 1. Bring up the stack
make compose

# 2. Define your foods once
myfitnesspal food add --name "Chicken Breast" --calories 165 --protein 31 --carbs 0 --fat 3.6
myfitnesspal food add --name "Oatmeal" --calories 150 --protein 5 --carbs 27 --fat 2.5

# 3. Check what you've defined
myfitnesspal food ls

# 4. Log what you eat through the day
myfitnesspal log add --food-id 1 --multiplier 1.5
myfitnesspal log add --food-id 2

# 5. Check today's totals
myfitnesspal log today

# 6. Clean up a mistaken entry
myfitnesspal log rm 8
```

---

## 7. Notes & current limitations

These reflect the current state of the project (per its own README) and are useful context when using the CLI:

- **Single user only** — there is no concept of accounts or auth.
- **No meals** — food is logged per-day, not grouped into breakfast/lunch/dinner.
- **No other tracked metrics** — water, steps, and supplements are out of scope.
- **Dates** must be `YYYY-MM-DD`; omitting `--date` on `log add`/`log today` defaults to the current day.
- All commands require the gRPC server (and its Postgres backend) to be reachable at `--addr` — there's no offline/local mode.
