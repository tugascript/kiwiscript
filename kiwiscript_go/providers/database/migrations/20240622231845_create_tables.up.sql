-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: PostgreSQL
-- Generated at: 2024-06-25T07:43:46.448Z

CREATE TABLE "users" (
  "id" serial PRIMARY KEY,
  "first_name" varchar(50) NOT NULL,
  "last_name" varchar(50) NOT NULL,
  "location" varchar(3) NOT NULL,
  "email" varchar(250) NOT NULL,
  "birth_date" date NOT NULL,
  "version" smallint NOT NULL DEFAULT 1,
  "is_admin" boolean NOT NULL DEFAULT false,
  "is_staff" boolean NOT NULL DEFAULT false,
  "is_confirmed" boolean NOT NULL,
  "password" text,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "auth_providers" (
  "id" serial PRIMARY KEY,
  "email" varchar(250) NOT NULL,
  "provider" varchar(8) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "languages" (
  "id" serial PRIMARY KEY,
  "name" varchar(50) UNIQUE NOT NULL,
  "icon" text NOT NULL,
  "author_id" int NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "categories" (
  "id" serial PRIMARY KEY,
  "title" varchar(100) NOT NULL,
  "slug" varchar(100) NOT NULL,
  "description" text NOT NULL,
  "author_id" int NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "series" (
  "id" serial PRIMARY KEY,
  "title" varchar(100) NOT NULL,
  "slug" varchar(100) NOT NULL,
  "description" text NOT NULL,
  "parts_count" smallint NOT NULL DEFAULT 0,
  "lectures_count" smallint NOT NULL DEFAULT 0,
  "total_duration_seconds" int NOT NULL DEFAULT 0,
  "review_avg" smallint NOT NULL DEFAULT 0,
  "review_count" int NOT NULL DEFAULT 0,
  "is_published" boolean NOT NULL DEFAULT false,
  "author_id" int NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "language_categories" (
  "language_id" int NOT NULL,
  "category_id" int NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("language_id", "category_id")
);

CREATE TABLE "series_categories" (
  "series_id" int NOT NULL,
  "category_id" int NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("series_id", "category_id")
);

CREATE TABLE "series_parts" (
  "id" serial PRIMARY KEY,
  "title" text NOT NULL,
  "series_id" int NOT NULL,
  "description" text NOT NULL,
  "position" smallint NOT NULL,
  "lectures_count" smallint NOT NULL DEFAULT 0,
  "total_duration_seconds" int NOT NULL DEFAULT 0,
  "is_published" boolean NOT NULL DEFAULT false,
  "author_id" int NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "lectures" (
  "id" serial PRIMARY KEY,
  "title" text NOT NULL,
  "video" varchar(250) NOT NULL,
  "duration_seconds" int NOT NULL,
  "position" smallint NOT NULL,
  "description" text NOT NULL,
  "is_published" boolean NOT NULL DEFAULT false,
  "comments_count" int NOT NULL DEFAULT 0,
  "author_id" int NOT NULL,
  "series_id" int NOT NULL,
  "series_part_id" int NOT NULL,
  "language_id" int NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "series_progress" (
  "id" serial PRIMARY KEY,
  "user_id" int NOT NULL,
  "series_id" int NOT NULL,
  "language_id" int NOT NULL,
  "lectures_count" smallint NOT NULL DEFAULT 0,
  "parts_count" smallint NOT NULL DEFAULT 0,
  "is_completed" boolean NOT NULL DEFAULT false,
  "completed_at" timestamp,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "series_part_progress" (
  "id" serial PRIMARY KEY,
  "user_id" int NOT NULL,
  "series_part_id" int NOT NULL,
  "language_id" int NOT NULL,
  "series_progress_id" int NOT NULL,
  "lectures_count" smallint NOT NULL DEFAULT 0,
  "is_completed" boolean NOT NULL DEFAULT false,
  "completed_at" timestamp,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "lecture_progress" (
  "id" serial PRIMARY KEY,
  "user_id" int NOT NULL,
  "language_id" int NOT NULL,
  "lecture_id" int NOT NULL,
  "series_progress_id" int NOT NULL,
  "series_part_progress_id" int NOT NULL,
  "is_completed" boolean NOT NULL DEFAULT false,
  "completed_at" timestamp,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "series_review" (
  "id" serial PRIMARY KEY,
  "author_id" int NOT NULL,
  "series_id" int NOT NULL,
  "language_id" int NOT NULL,
  "rating" smallint NOT NULL,
  "review" text,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "lecture_comments" (
  "id" serial PRIMARY KEY,
  "author_id" int NOT NULL,
  "lecture_id" int NOT NULL,
  "comment" text NOT NULL,
  "replies_count" int NOT NULL DEFAULT 0,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "comment_replies" (
  "id" serial PRIMARY KEY,
  "author_id" int NOT NULL,
  "comment_id" int NOT NULL,
  "reply" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "certificates" (
  "id" serial PRIMARY KEY,
  "reference" uuid NOT NULL,
  "user_id" int NOT NULL,
  "language_id" int NOT NULL,
  "series_id" int NOT NULL,
  "completed_at" timestamp NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "donations" (
  "id" serial PRIMARY KEY,
  "user_id" int NOT NULL,
  "amount" bigint NOT NULL,
  "currency" varchar(3) NOT NULL,
  "recurring" boolean NOT NULL DEFAULT false,
  "recurring_ref" varchar(250),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "payments" (
  "id" serial PRIMARY KEY,
  "payment_ref" varchar(250) NOT NULL,
  "user_id" int NOT NULL,
  "donation_id" int NOT NULL,
  "amount" bigint NOT NULL,
  "currency" varchar(3) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE UNIQUE INDEX "users_email_unique_idx" ON "users" ("email");

CREATE INDEX "auth_providers_email_idx" ON "auth_providers" ("email");

CREATE UNIQUE INDEX "auth_providers_email_provider_unique_idx" ON "auth_providers" ("email", "provider");

CREATE UNIQUE INDEX "languages_name_unique_idx" ON "languages" ("name");

CREATE UNIQUE INDEX "categories_title_unique_idx" ON "categories" ("title");

CREATE UNIQUE INDEX "categories_slug_unique_idx" ON "categories" ("slug");

CREATE UNIQUE INDEX "series_title_unique_idx" ON "series" ("title");

CREATE UNIQUE INDEX "series_slug_unique_idx" ON "series" ("slug");

CREATE INDEX "series_is_published_idx" ON "series" ("is_published");

CREATE UNIQUE INDEX "series_parts_title_series_id_unique_idx" ON "series_parts" ("title", "series_id");

CREATE UNIQUE INDEX "series_parts_series_id_position_unique_idx" ON "series_parts" ("series_id", "position");

CREATE INDEX "series_parts_series_id_idx" ON "series_parts" ("series_id");

CREATE INDEX "series_parts_author_id_idx" ON "series_parts" ("author_id");

CREATE INDEX "series_parts_is_published_idx" ON "series_parts" ("is_published");

CREATE INDEX "series_parts_position_idx" ON "series_parts" ("position");

CREATE UNIQUE INDEX "lectures_title_series_id_series_part_id_unique_idx" ON "lectures" ("title", "series_id", "series_part_id");

CREATE UNIQUE INDEX "lectures_series_part_id_position_unique_idx" ON "lectures" ("series_part_id", "position");

CREATE INDEX "lectures_series_id_idx" ON "lectures" ("series_id");

CREATE INDEX "lectures_language_id_idx" ON "lectures" ("language_id");

CREATE INDEX "lectures_series_part_id_idx" ON "lectures" ("series_part_id");

CREATE INDEX "lectures_author_id_idx" ON "lectures" ("author_id");

CREATE INDEX "lectures_is_listed_idx" ON "lectures" ("is_published");

CREATE INDEX "lectures_position_idx" ON "lectures" ("position");

CREATE UNIQUE INDEX "series_progress_user_id_series_id_language_id_unique_idx" ON "series_progress" ("user_id", "series_id", "language_id");

CREATE INDEX "series_progress_user_id_idx" ON "series_progress" ("user_id");

CREATE INDEX "series_progress_series_id_idx" ON "series_progress" ("series_id");

CREATE INDEX "series_progress_language_id_idx" ON "series_progress" ("language_id");

CREATE UNIQUE INDEX "series_part_progress_user_id_series_part_id_language_id_unique_idx" ON "series_part_progress" ("user_id", "series_part_id", "language_id");

CREATE UNIQUE INDEX "series_part_progress_user_id_series_progress_id_language_id_unique_idx" ON "series_part_progress" ("user_id", "series_progress_id", "language_id");

CREATE INDEX "series_part_progress_user_id_idx" ON "series_part_progress" ("user_id");

CREATE INDEX "series_part_progress_series_part_id_idx" ON "series_part_progress" ("series_part_id");

CREATE INDEX "series_part_progress_series_progress_id_idx" ON "series_part_progress" ("series_progress_id");

CREATE INDEX "series_part_progress_language_id_idx" ON "series_part_progress" ("language_id");

CREATE UNIQUE INDEX "lecture_progress_user_id_lecture_id_language_id_unique_idx" ON "lecture_progress" ("user_id", "lecture_id", "language_id");

CREATE UNIQUE INDEX "lecture_progress_user_id_series_part_progress_id_language_id_unique_idx" ON "lecture_progress" ("user_id", "series_part_progress_id", "language_id");

CREATE INDEX "lecture_progress_user_id_idx" ON "lecture_progress" ("user_id");

CREATE INDEX "lecture_progress_lecture_id_idx" ON "lecture_progress" ("lecture_id");

CREATE INDEX "lecture_progress_series_progress_id_idx" ON "lecture_progress" ("series_progress_id");

CREATE INDEX "lecture_progress_series_part_progress_id_idx" ON "lecture_progress" ("series_part_progress_id");

CREATE INDEX "lecture_progress_language_id_idx" ON "lecture_progress" ("language_id");

CREATE UNIQUE INDEX "series_review_user_id_series_id_language_id_unique_idx" ON "series_review" ("author_id", "series_id", "language_id");

CREATE INDEX "series_review_author_id_idx" ON "series_review" ("author_id");

CREATE INDEX "series_review_series_id_idx" ON "series_review" ("series_id");

CREATE INDEX "series_review_language_id_idx" ON "series_review" ("language_id");

CREATE INDEX "lecture_comments_author_id_idx" ON "lecture_comments" ("author_id");

CREATE INDEX "comment_replies_author_id_idx" ON "comment_replies" ("author_id");

CREATE INDEX "comment_replies_comment_id_idx" ON "comment_replies" ("comment_id");

CREATE UNIQUE INDEX "certificates_user_id_language_id_series_id_unique_idx" ON "certificates" ("user_id", "language_id", "series_id");

CREATE UNIQUE INDEX "certificates_reference_unique_idx" ON "certificates" ("reference");

CREATE INDEX "certificates_user_id_idx" ON "certificates" ("user_id");

CREATE INDEX "certificates_language_id_idx" ON "certificates" ("language_id");

CREATE INDEX "certificates_series_id_idx" ON "certificates" ("series_id");

CREATE INDEX "donations_user_id_idx" ON "donations" ("user_id");

CREATE INDEX "payments_user_id_idx" ON "payments" ("user_id");

CREATE INDEX "payments_donation_id_idx" ON "payments" ("donation_id");

ALTER TABLE "auth_providers" ADD FOREIGN KEY ("email") REFERENCES "users" ("email") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "languages" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "categories" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "language_categories" ADD FOREIGN KEY ("language_id") REFERENCES "languages" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "language_categories" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_categories" ADD FOREIGN KEY ("series_id") REFERENCES "series" ("id");

ALTER TABLE "series_categories" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

ALTER TABLE "series_parts" ADD FOREIGN KEY ("series_id") REFERENCES "series" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_parts" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lectures" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lectures" ADD FOREIGN KEY ("series_id") REFERENCES "series" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lectures" ADD FOREIGN KEY ("series_part_id") REFERENCES "series_parts" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lectures" ADD FOREIGN KEY ("language_id") REFERENCES "languages" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_progress" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_progress" ADD FOREIGN KEY ("series_id") REFERENCES "series" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_progress" ADD FOREIGN KEY ("language_id") REFERENCES "languages" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_part_progress" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_part_progress" ADD FOREIGN KEY ("series_part_id") REFERENCES "series_parts" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_part_progress" ADD FOREIGN KEY ("language_id") REFERENCES "languages" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_part_progress" ADD FOREIGN KEY ("series_progress_id") REFERENCES "series_progress" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lecture_progress" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lecture_progress" ADD FOREIGN KEY ("language_id") REFERENCES "languages" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lecture_progress" ADD FOREIGN KEY ("lecture_id") REFERENCES "lectures" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lecture_progress" ADD FOREIGN KEY ("series_progress_id") REFERENCES "series_progress" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lecture_progress" ADD FOREIGN KEY ("series_part_progress_id") REFERENCES "series_part_progress" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_review" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_review" ADD FOREIGN KEY ("series_id") REFERENCES "series" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_review" ADD FOREIGN KEY ("language_id") REFERENCES "languages" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lecture_comments" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lecture_comments" ADD FOREIGN KEY ("lecture_id") REFERENCES "lectures" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "certificates" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "certificates" ADD FOREIGN KEY ("language_id") REFERENCES "languages" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "certificates" ADD FOREIGN KEY ("series_id") REFERENCES "series" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "donations" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "payments" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "payments" ADD FOREIGN KEY ("donation_id") REFERENCES "donations" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
