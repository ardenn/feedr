BEGIN;
ALTER TABLE IF EXISTS feeds DROP COLUMN rhash;
COMMIT;