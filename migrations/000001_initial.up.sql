BEGIN;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE IF NOT EXISTS "users" (
    "id" uuid DEFAULT gen_random_uuid(),
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz,
    "user_id" bigint NOT NULL,
    "chat_id" bigint NOT NULL,
    "is_bot" boolean NOT NULL DEFAULT false,
    "first_name" text,
    "last_name" text,
    "username" text,
    "last_fetch" timestamptz,
    PRIMARY KEY ("id"),
    UNIQUE ("chat_id"),
    UNIQUE ("user_id")
);
CREATE TABLE IF NOT EXISTS "feeds" (
    "id" uuid DEFAULT gen_random_uuid(),
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL,
    "user_id" uuid NOT NULL,
    "is_rss" boolean NOT NULL DEFAULT true,
    "link" text NOT NULL,
    "name" text NOT NULL,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);
COMMIT;