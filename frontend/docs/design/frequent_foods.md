# Plan: Frequent vs. Rare Foods

## Problem

The `foods` table has ~30 rows: ~8 eaten almost every day, ~20 kept around for
occasional use. `ListFoods` returns all of them with no way to distinguish
the two groups, so the CLI (and eventually the frontend) always renders the
full list. There's no way to ask for just "the foods I actually use."

## Goal

`ListFoods` can be filtered down to just the frequently-used foods (or just
the rest), while still supporting "give me everything" as the default. Foods
also need a way to be marked frequent/not-frequent.

## Non-goals (for this pass)

- Auto-detecting "frequent" from `log_entries` history (e.g. "logged 5+ times
  in the last 2 weeks"). That's a nice future enhancement but adds real
  complexity (what window? what threshold? recompute when?) for a
  30-item list. Starting with a manual flag the user sets once and rarely
  touches is simpler and fully sufficient at this scale.
- Any frontend work — there's no frontend code yet beyond generated stubs.
- Renaming/restructuring `Food` beyond adding the one field.

If usage grows a lot, the computed approach can be layered on later without
breaking this design (see "Future option" at the bottom).

## Design

Add a boolean column `is_frequent` on `foods`, defaulting to `false`. Expose
it on the `Food` proto message, let `CreateFood` optionally set it, add a
dedicated RPC to flip it later, and let `ListFoods` take an optional filter.

### 1. Migration

New goose migration, e.g. `db/migrations/<ts>_add_is_frequent_to_foods.sql`:

```sql
-- +goose Up
ALTER TABLE foods ADD COLUMN is_frequent BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE foods DROP COLUMN is_frequent;
```

After this lands, you'd run something like `food set-frequent <id> true` for
your ~8 daily foods (see CLI section) as a one-time backfill.

### 2. sqlc queries (`db/queries/foods.sql`)

Use `sqlc.narg` so `ListFoods` can take an optional filter without needing
two separate query names:

```sql
-- name: CreateFood :one
INSERT INTO foods (name, calories, protein_g, carbs_g, fat_g, is_frequent)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetFood :one
SELECT * FROM foods WHERE id = $1;

-- name: ListFoods :many
SELECT * FROM foods
WHERE sqlc.narg('is_frequent')::boolean IS NULL
   OR is_frequent = sqlc.narg('is_frequent')
ORDER BY id;

-- name: SetFoodFrequent :one
UPDATE foods SET is_frequent = $2 WHERE id = $1
RETURNING *;

-- name: DeleteFood :exec
DELETE FROM foods WHERE id = $1;
```

`ListFoods` generated Go signature becomes something like
`ListFoods(ctx, pgtype.Bool)` — pass an invalid/null `pgtype.Bool` to mean
"no filter."

### 3. Proto (`proto/food/v1/food.proto`)

```protobuf
message Food {
  int32 id = 1;
  string name = 2;
  double calories = 3;
  double protein_g = 4;
  double carbs_g = 5;
  double fat_g = 6;
  bool is_frequent = 7;
}

message CreateFoodRequest {
  string name = 1;
  double calories = 2;
  double protein_g = 3;
  double carbs_g = 4;
  double fat_g = 5;
  bool is_frequent = 6; // optional at creation time, defaults to false
}

message ListFoodsRequest {
  // unset = all foods, true = frequent only, false = rare only
  optional bool is_frequent = 1;
}
message ListFoodsResponse {
  repeated Food foods = 1;
}

message SetFoodFrequentRequest {
  int32 id = 1;
  bool is_frequent = 2;
}
message SetFoodFrequentResponse {
  Food food = 1;
}

service FoodService {
  rpc CreateFood(CreateFoodRequest) returns (CreateFoodResponse);
  rpc DeleteFood(DeleteFoodRequest) returns (DeleteFoodResponse);
  rpc ListFoods(ListFoodsRequest) returns (ListFoodsResponse);
  rpc GetFood(GetFoodRequest) returns (GetFoodResponse);
  rpc SetFoodFrequent(SetFoodFrequentRequest) returns (SetFoodFrequentResponse);
}
```

Using `optional bool` on `ListFoodsRequest` (rather than a plain `bool`)
matters: a plain `bool` can't distinguish "not set, show me everything" from
"explicitly false, show me the rare ones," since both would default to
`false` on the wire. `optional` generates a presence-checkable field
(`req.Msg.IsFrequent` becomes a `*bool` in Go).

Then regenerate stubs: `buf generate` (per the existing `buf.gen.yaml`).

### 4. Service layer (`internal/service/food.go`)

```go
func (s *FoodService) ListFoods(ctx context.Context, req *connect.Request[foodv1.ListFoodsRequest]) (*connect.Response[foodv1.ListFoodsResponse], error) {
	var filter pgtype.Bool
	if req.Msg.IsFrequent != nil {
		filter = pgtype.Bool{Bool: *req.Msg.IsFrequent, Valid: true}
	}
	rows, err := s.q.ListFoods(ctx, filter)
	if err != nil {
		return nil, err
	}
	foods := make([]*foodv1.Food, len(rows))
	for i, row := range rows {
		foods[i] = toProtoFood(row)
	}
	return connect.NewResponse(&foodv1.ListFoodsResponse{Foods: foods}), nil
}

func (s *FoodService) SetFoodFrequent(ctx context.Context, req *connect.Request[foodv1.SetFoodFrequentRequest]) (*connect.Response[foodv1.SetFoodFrequentResponse], error) {
	row, err := s.q.SetFoodFrequent(ctx, db.SetFoodFrequentParams{
		ID:         req.Msg.Id,
		IsFrequent: req.Msg.IsFrequent,
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&foodv1.SetFoodFrequentResponse{Food: toProtoFood(row)}), nil
}
```

Also add `IsFrequent: row.IsFrequent` to `toProtoFood`, and pass
`req.Msg.IsFrequent` through in `CreateFood`.

### 5. CLI (`cmd/cli/cmd/food.go`)

- `food ls` — no change in default behavior, still lists everything.
- `food ls --frequent` — only frequent foods.
- `food ls --rare` — only non-frequent foods (make `--frequent`/`--rare`
  mutually exclusive, e.g. with cobra's `MarkFlagsMutuallyExclusive`).
- New command: `food set-frequent <id> <true|false>` to flip the flag.
- `food add` gets an optional `--frequent` flag (defaults to false) so new
  daily staples can be marked frequent right away instead of a two-step
  add-then-flag.

### 6. Backfill

Once deployed, run `food ls` to find your ~8 daily foods' IDs and call
`food set-frequent <id> true` on each. One-time, no data migration script
needed since the default (`false`) is already correct for the other ~20.

## Rollout order

1. Migration (additive, safe to deploy alone — `is_frequent` defaults to
   `false`, nothing breaks).
2. Proto changes + regenerate stubs.
3. sqlc query changes + regenerate.
4. Service layer changes.
5. CLI changes.
6. Backfill your 8 daily foods.

Each step after the migration only compiles/works if the prior step's
generated code is in place, so this is naturally sequential — no need to
flag-gate or dual-write anything given this is a single-user local tool.

## Testing checklist

- `ListFoods` with no filter still returns all ~30 foods.
- `ListFoods` with `is_frequent=true` returns only the flagged ones.
- `ListFoods` with `is_frequent=false` returns only the rest.
- `SetFoodFrequent` flips the flag and the change is visible in a subsequent
  `ListFoods` call.
- `CreateFood` with `is_frequent` unset defaults to `false` (check the DB
  default kicks in as expected via sqlc's generated struct zero value).
