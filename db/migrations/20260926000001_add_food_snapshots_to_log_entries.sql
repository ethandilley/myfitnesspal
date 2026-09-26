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