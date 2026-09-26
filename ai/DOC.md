# Denormalize Food Snapshots into `log_entries`

## Problem

`DeleteFood` cascades to `log_entries` because of:

```sql
food_id INTEGER NOT NULL REFERENCES foods(id) ON DELETE CASCADE
```

Deleting a food destroys all prior daily logs and macro history for that food.

## Desired Behavior

- `DeleteFood` sets `log_entries.food_id` to `NULL` instead of deleting rows.
- Each `log_entries` row snapshots the food name and macros at insert time.
- Macro totals for any date are computed from the snapshot columns, not a join to `foods`.
- API consumers receive snapshot fields on `LogEntry` so deleted-food entries remain readable.

---

## 1. Database Migration

Create a new goose migration file in `db/migrations/` with a timestamp after `20260915180610`, e.g.:

```
db/migrations/20260926000001_add_food_snapshots_to_log_entries.sql
```

### Up

```sql
-- +goose Up
ALTER TABLE log_entries
  ADD COLUMN food_name TEXT NOT NULL DEFAULT '',
  ADD COLUMN calories NUMERIC NOT NULL DEFAULT 0,
  ADD COLUMN protein_g NUMERIC NOT NULL DEFAULT 0,
  ADD COLUMN carbs_g NUMERIC NOT NULL DEFAULT 0,
  ADD COLUMN fat_g NUMERIC NOT NULL DEFAULT 0;

UPDATE log_entries le
SET
  food_name = f.name,
  calories  = f.calories,
  protein_g = f.protein_g,
  carbs_g   = f.carbs_g,
  fat_g     = f.fat_g
FROM foods f
WHERE f.id = le.food_id;

ALTER TABLE log_entries
  ALTER COLUMN food_id DROP NOT NULL;

ALTER TABLE log_entries
  DROP CONSTRAINT log_entries_food_id_fkey;

ALTER TABLE log_entries
  ADD CONSTRAINT log_entries_food_id_fkey
  FOREIGN KEY (food_id) REFERENCES foods(id) ON DELETE SET NULL;
```

### Down

```sql
-- +goose Down
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM log_entries WHERE food_id IS NULL) THEN
    RAISE EXCEPTION 'Refusing to roll back: log_entries contains food_id IS NULL';
  END IF;
END
$$;

ALTER TABLE log_entries DROP CONSTRAINT log_entries_food_id_fkey;
ALTER TABLE log_entries ALTER COLUMN food_id SET NOT NULL;
ALTER TABLE log_entries
  ADD CONSTRAINT log_entries_food_id_fkey
  FOREIGN KEY (food_id) REFERENCES foods(id) ON DELETE CASCADE;
ALTER TABLE log_entries
  DROP COLUMN food_name,
  DROP COLUMN calories,
  DROP COLUMN protein_g,
  DROP COLUMN carbs_g,
  DROP COLUMN fat_g;
```

Apply with:

```bash
make migrate
```

---

## 2. sqlc Query Changes

Edit `db/queries/logs.sql`.

### `CreateLogEntry`

Replace the existing insert with a `SELECT … FROM foods` so the food must exist and snapshot columns are populated atomically:

```sql
-- name: CreateLogEntry :one
INSERT INTO log_entries (
  food_id,
  multiplier,
  logged_at,
  food_name,
  calories,
  protein_g,
  carbs_g,
  fat_g
)
SELECT
  $1,
  $2,
  COALESCE(sqlc.narg('logged_at')::date, CURRENT_DATE),
  f.name,
  f.calories,
  f.protein_g,
  f.carbs_g,
  f.fat_g
FROM foods f
WHERE f.id = $1
RETURNING *;
```

Parameter order stays: `$1` = food_id, `$2` = multiplier, optional `logged_at`.

### `GetMacroTotalsByDate`

Replace the `JOIN foods` with direct aggregation on the snapshot columns:

```sql
-- name: GetMacroTotalsByDate :one
SELECT
  COALESCE(SUM(l.calories * l.multiplier), 0)::numeric   AS calories,
  COALESCE(SUM(l.protein_g * l.multiplier), 0)::numeric  AS protein_g,
  COALESCE(SUM(l.carbs_g * l.multiplier), 0)::numeric    AS carbs_g,
  COALESCE(SUM(l.fat_g * l.multiplier), 0)::numeric      AS fat_g
FROM log_entries l
WHERE l.logged_at = $1;
```

### `ListLogEntries` / `ListLogEntriesByDate` / `GetLogEntry` / `DeleteLogEntry`

No changes needed. They use `SELECT *` or simple DML and will pick up the new columns automatically after regeneration.

### `db/queries/foods.sql`

No changes. `DeleteFood` stays `DELETE FROM foods WHERE id = $1;` — the FK change handles the rest.

### Regenerate

```bash
sqlc generate
```

After regeneration, inspect `internal/db/models.go` and `internal/db/logs.sql.go`. Expected changes:

- `LogEntry.FoodID` becomes `pgtype.Int4` (nullable).
- `LogEntry` gains `FoodName string`, `Calories`, `ProteinG`, `CarbsG`, `FatG` (all `pgtype.Numeric` except `FoodName`).
- `CreateLogEntryParams.FoodID` may become `pgtype.Int4`.

---

## 3. Proto Contract

Edit `proto/log/v1/log.proto`. Add five fields to `LogEntry` using unused tags 5–9:

```proto
message LogEntry {
  int32 id = 1;
  int32 food_id = 2;        // 0 means the food was deleted
  double multiplier = 3;
  string logged_at = 4;     // YYYY-MM-DD
  string food_name = 5;
  double calories = 6;
  double protein_g = 7;
  double carbs_g = 8;
  double fat_g = 9;
}
```

`food_id` stays a plain `int32` (no `google.protobuf.Int32Value`). A value of `0` signals the food no longer exists.

Regenerate Go and TypeScript clients:

```bash
buf generate
```

---

## 4. Service Layer

Edit `internal/service/log.go`.

### Imports

Add if not already present:

```go
"database/sql"
"errors"
"fmt"
```

(`fmt` may already be used; check.)

### `CreateLogEntry`

Handle the case where the food does not exist. The `SELECT … FROM foods` query returns zero rows, so sqlc/pgx surfaces `sql.ErrNoRows`:

```go
row, err := s.q.CreateLogEntry(ctx, db.CreateLogEntryParams{
    FoodID:     pgtype.Int4{Int32: msg.GetFoodId(), Valid: true},
    Multiplier: floatToNumeric(msg.Multiplier),
    LoggedAt:   loggedAt,
})
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("food %d not found", msg.GetFoodId()))
    }
    return nil, connect.NewError(connect.CodeInternal, err)
}
```

> If `sqlc generate` leaves `CreateLogEntryParams.FoodID` as `int32` (not `pgtype.Int4`), pass `msg.GetFoodId()` directly instead.

### `toProtoLogEntry`

Map the new snapshot fields and the nullable `FoodID`:

```go
func toProtoLogEntry(row db.LogEntry) *logv1.LogEntry {
    foodID := int32(0)
    if row.FoodID.Valid {
        foodID = row.FoodID.Int32
    }
    return &logv1.LogEntry{
        Id:         row.ID,
        FoodId:     foodID,
        Multiplier: numericToFloat(row.Multiplier),
        LoggedAt:   dateToString(row.LoggedAt),
        FoodName:   row.FoodName,
        Calories:   numericToFloat(row.Calories),
        ProteinG:   numericToFloat(row.ProteinG),
        CarbsG:     numericToFloat(row.CarbsG),
        FatG:       numericToFloat(row.FatG),
    }
}
```

`numericToFloat` and `dateToString` already exist in the package.

---

## 5. CLI

Edit `cmd/cli/cmd/log.go`.

In `log ls` and `log today`, when printing each entry use `e.FoodName` if non-empty, otherwise fall back to `food #<id>`:

```go
name := e.GetFoodName()
if name == "" {
    name = fmt.Sprintf("food #%d", e.GetFoodId())
}
```

The generated client types pick up the new fields after `buf generate` + `go build ./...`.

---

## 6. Frontend

Edit `frontend/index.html` (and any generated-client usage).

- Read `entry.foodName`, `entry.calories`, `entry.proteinG`, `entry.carbsG`, `entry.fatG` from `LogEntry` objects.
- Fall back to `Food #<foodId>` when `foodName` is empty and `foodId` is `0`.
- Daily totals already come from the server; no client-side join needed.

Regenerate TS clients with `buf generate` before building the frontend.

---

## 7. Verification Checklist

1. **Migration**
   ```bash
   make migrate
   psql "$DB_URL" -c '\d log_entries'
   ```
   Confirm five new columns, nullable `food_id`, and `ON DELETE SET NULL`.

2. **Codegen**
   ```bash
   sqlc generate
   buf generate
   go build ./...
   ```

3. **Unit / smoke test**
   ```bash
   make compose
   go run cmd/cli/main.go food add --name "Test Food" --calories 100 --protein 10 --carbs 5 --fat 2
   go run cmd/cli/main.go log add --food-id 1 --multiplier 1
   go run cmd/cli/main.go log ls
   go run cmd/cli/main.go food rm 1
   go run cmd/cli/main.go log ls   # should still show "Test Food" via snapshot
   ```

4. **Macro totals**
   After deleting the food, `log today` (or the frontend) should still report correct totals for the snapshot row.

5. **Rollback (optional, dev only)**
   ```bash
   goose -dir db/migrations postgres "$DB_URL" down 1
   ```
   Will abort if any `log_entries.food_id IS NULL`.

---

## 8. Files Touched (summary)

| File | Change |
|---|---|
| `db/migrations/20260926000001_*.sql` | New migration |
| `db/queries/logs.sql` | `CreateLogEntry`, `GetMacroTotalsByDate` |
| `internal/db/models.go` | Regenerated |
| `internal/db/logs.sql.go` | Regenerated |
| `proto/log/v1/log.proto` | Five new `LogEntry` fields |
| `gen/proto/log/v1/*.pb.go` | Regenerated |
| `frontend/src/gen/**` | Regenerated |
| `internal/service/log.go` | `CreateLogEntry` error handling, `toProtoLogEntry` |
| `cmd/cli/cmd/log.go` | Display snapshot `FoodName` |
| `frontend/index.html` | Display snapshot fields |
