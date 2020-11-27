BEGIN;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE IF NOT EXISTS "users" (
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz,
    "id" bigserial NOT NULL,
    "is_bot" boolean NOT NULL DEFAULT false,
    "first_name" text,
    "last_name" text,
    "username" text,
    "last_fetch" timestamptz,
    PRIMARY KEY ("id")
);
CREATE TABLE IF NOT EXISTS "feeds" (
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz,
    "id" uuid DEFAULT gen_random_uuid(),
    "user_id" bigint NOT NULL,
    "is_rss" boolean NOT NULL DEFAULT true,
    "link" text NOT NULL,
    "name" text NOT NULL,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);
CREATE INDEX ix_feeds_user_id ON public.feeds USING btree (user_id);
COMMIT;