CREATE TABLE IF NOT EXISTS "posts" (
"id" bigserial PRIMARY KEY,
"title" text NOT NULL,
"post_text" text NOT NULL,
"img" text NOT NULL,
"read_time" integer NOT NULL,
"liked_by" integer[],
"user_id" bigint NOT NULL,
"user_name" text NOT NULL,
"version" integer NOT NULL DEFAULT 1,
"created_at" timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "comments" (
"id" bigserial PRIMARY KEY,
"text" text NOT NULL,
"user_id" bigint NOT NULL,
"user_name" text NOT NULL,
"post_id" bigint NOT NULL,
"created_at" timestamp(0) with time zone NOT NULL DEFAULT NOW()
);