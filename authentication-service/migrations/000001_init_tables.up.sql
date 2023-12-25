CREATE TABLE IF NOT EXISTS "users" (
"id" bigserial PRIMARY KEY,
"name" text NOT NULL,
"email" citext UNIQUE NOT NULL,
"password_hash" bytea NOT NULL,
"bio" text,
"avatar" text,
"active" BOOLEAN NOT NULL,
"admin" BOOLEAN NOT NULL,
"version" integer NOT NULL DEFAULT 1,
"created_at" timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "tokens" (
"hash" bytea PRIMARY KEY,
"user_id" bigint NOT NULL REFERENCES users ON DELETE CASCADE,
"expiry" timestamp(0) with time zone NOT NULL,
"scope" text NOT NULL
);