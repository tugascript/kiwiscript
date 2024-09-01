-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: PostgreSQL
-- Generated at: 2024-08-31T07:58:39.473Z

CREATE TABLE "users" (
  "id" serial PRIMARY KEY,
  "first_name" varchar(50) NOT NULL,
  "last_name" varchar(50) NOT NULL,
  "location" varchar(3) NOT NULL,
  "email" varchar(250) NOT NULL,
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
  "name" varchar(50) NOT NULL,
  "slug" varchar(50) NOT NULL,
  "icon" text NOT NULL,
  "series_count" smallint NOT NULL DEFAULT 0,
  "author_id" int NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "series" (
  "id" serial PRIMARY KEY,
  "title" varchar(100) NOT NULL,
  "slug" varchar(100) NOT NULL,
  "description" text NOT NULL,
  "sections_count" smallint NOT NULL DEFAULT 0,
  "lessons_count" smallint NOT NULL DEFAULT 0,
  "watch_time_seconds" int NOT NULL DEFAULT 0,
  "read_time_seconds" int NOT NULL DEFAULT 0,
  "is_published" boolean NOT NULL DEFAULT false,
  "language_slug" varchar(50) NOT NULL,
  "author_id" int NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "series_pictures" (
  "id" uuid PRIMARY KEY,
  "series_id" int NOT NULL,
  "author_id" int NOT NULL,
  "ext" varchar(10) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "sections" (
  "id" serial PRIMARY KEY,
  "title" varchar(250) NOT NULL,
  "language_slug" varchar(50) NOT NULL,
  "series_slug" varchar(100) NOT NULL,
  "description" text NOT NULL,
  "position" smallint NOT NULL,
  "lessons_count" smallint NOT NULL DEFAULT 0,
  "watch_time_seconds" int NOT NULL DEFAULT 0,
  "read_time_seconds" int NOT NULL DEFAULT 0,
  "is_published" boolean NOT NULL DEFAULT false,
  "author_id" int NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "lessons" (
  "id" serial PRIMARY KEY,
  "title" varchar(250) NOT NULL,
  "position" smallint NOT NULL,
  "is_published" boolean NOT NULL DEFAULT false,
  "watch_time_seconds" int NOT NULL DEFAULT 0,
  "read_time_seconds" int NOT NULL DEFAULT 0,
  "author_id" int NOT NULL,
  "language_slug" varchar(50) NOT NULL,
  "series_slug" varchar(100) NOT NULL,
  "section_id" int NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "lesson_articles" (
  "id" serial PRIMARY KEY,
  "lesson_id" int NOT NULL,
  "author_id" int NOT NULL,
  "content" text NOT NULL,
  "read_time_seconds" int NOT NULL DEFAULT 0,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "lesson_videos" (
  "id" serial PRIMARY KEY,
  "lesson_id" int NOT NULL,
  "author_id" int NOT NULL,
  "url" varchar(250) NOT NULL,
  "watch_time_seconds" int NOT NULL DEFAULT 0,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "lesson_files" (
  "id" uuid PRIMARY KEY,
  "lesson_id" int NOT NULL,
  "author_id" int NOT NULL,
  "ext" varchar(10) NOT NULL,
  "name" varchar(250) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "language_progress" (
  "id" serial PRIMARY KEY,
  "user_id" int NOT NULL,
  "language_slug" varchar(50) NOT NULL,
  "completed_series" smallint NOT NULL DEFAULT 0,
  "viewed_at" timestamp NOT NULL DEFAULT (now()),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "series_progress" (
  "id" serial PRIMARY KEY,
  "user_id" int NOT NULL,
  "series_slug" varchar(100) NOT NULL,
  "language_slug" varchar(50) NOT NULL,
  "language_progress_id" int NOT NULL,
  "completed_sections" smallint NOT NULL DEFAULT 0,
  "completed_lessons" smallint NOT NULL DEFAULT 0,
  "parts_count" smallint NOT NULL DEFAULT 0,
  "completed_at" timestamp,
  "viewed_at" timestamp NOT NULL DEFAULT (now()),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "section_progress" (
  "id" serial PRIMARY KEY,
  "user_id" int NOT NULL,
  "language_slug" varchar(50) NOT NULL,
  "series_slug" varchar(100) NOT NULL,
  "section_id" int NOT NULL,
  "language_progress_id" int NOT NULL,
  "series_progress_id" int NOT NULL,
  "completed_lessons" smallint NOT NULL DEFAULT 0,
  "completed_at" timestamp,
  "viewed_at" timestamp NOT NULL DEFAULT (now()),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "lesson_progress" (
  "id" serial PRIMARY KEY,
  "user_id" int NOT NULL,
  "language_slug" varchar(50) NOT NULL,
  "series_slug" varchar(100) NOT NULL,
  "section_id" int NOT NULL,
  "lesson_id" int NOT NULL,
  "language_progress_id" int NOT NULL,
  "series_progress_id" int NOT NULL,
  "section_progress_id" int NOT NULL,
  "completed_at" timestamp,
  "viewed_at" timestamp NOT NULL DEFAULT (now()),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "certificates" (
  "id" uuid PRIMARY KEY,
  "user_id" int NOT NULL,
  "series_title" varchar(100) NOT NULL,
  "lessons" smallint NOT NULL,
  "watch_time_seconds" int NOT NULL,
  "read_time_seconds" int NOT NULL,
  "language_slug" varchar(50) NOT NULL,
  "series_slug" varchar(100) NOT NULL,
  "completed_at" timestamp NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE UNIQUE INDEX "users_email_unique_idx" ON "users" ("email");

CREATE INDEX "users_is_staff_idx" ON "users" ("is_staff");

CREATE INDEX "auth_providers_email_idx" ON "auth_providers" ("email");

CREATE UNIQUE INDEX "auth_providers_email_provider_unique_idx" ON "auth_providers" ("email", "provider");

CREATE UNIQUE INDEX "languages_name_unique_idx" ON "languages" ("name");

CREATE UNIQUE INDEX "languages_slug_unique_idx" ON "languages" ("slug");

CREATE UNIQUE INDEX "series_title_unique_idx" ON "series" ("title");

CREATE UNIQUE INDEX "series_slug_unique_idx" ON "series" ("slug");

CREATE INDEX "series_is_published_idx" ON "series" ("is_published");

CREATE INDEX "series_language_slug_idx" ON "series" ("language_slug");

CREATE INDEX "series_slug_language_slug_idx" ON "series" ("slug", "language_slug");

CREATE INDEX "series_id_language_slug_idx" ON "series" ("id", "language_slug");

CREATE INDEX "series_id_language_slug_is_published_idx" ON "series" ("id", "language_slug", "is_published");

CREATE INDEX "series_slug_is_published_idx" ON "series" ("slug", "language_slug", "is_published");

CREATE INDEX "series_author_id_idx" ON "series" ("author_id");

CREATE UNIQUE INDEX "series_images_series_id_unique_idx" ON "series_pictures" ("series_id");

CREATE INDEX "series_images_author_id_idx" ON "series_pictures" ("author_id");

CREATE UNIQUE INDEX "sections_title_series_slug_unique_idx" ON "sections" ("title", "series_slug");

CREATE INDEX "sections_series_slug_position_unique_idx" ON "sections" ("series_slug", "position");

CREATE INDEX "sections_language_slug_series_slug_idx" ON "sections" ("language_slug", "series_slug");

CREATE INDEX "sections_language_slug_series_slug_is_published_idx" ON "sections" ("language_slug", "series_slug", "is_published");

CREATE INDEX "sections_language_slug_series_slug_id_idx" ON "sections" ("language_slug", "series_slug", "id");

CREATE INDEX "sections_language_slug_idx" ON "sections" ("language_slug");

CREATE INDEX "sections_series_slug_idx" ON "sections" ("series_slug");

CREATE INDEX "sections_author_id_idx" ON "sections" ("author_id");

CREATE INDEX "sections_is_published_idx" ON "sections" ("is_published");

CREATE INDEX "sections_position_idx" ON "sections" ("position");

CREATE UNIQUE INDEX "lessons_title_section_id_unique_idx" ON "lessons" ("title", "section_id");

CREATE INDEX "lessons_language_slug_series_slug_idx" ON "lessons" ("language_slug", "series_slug");

CREATE INDEX "lessons_language_slug_series_slug_is_published_idx" ON "lessons" ("language_slug", "series_slug", "is_published");

CREATE INDEX "lessons_language_slug_series_slug_id_idx" ON "lessons" ("language_slug", "series_slug", "id");

CREATE INDEX "lessons_language_slug_idx" ON "lessons" ("language_slug");

CREATE INDEX "lessons_series_slug_idx" ON "lessons" ("series_slug");

CREATE INDEX "lessons_section_id_idx" ON "lessons" ("section_id");

CREATE INDEX "lessons_author_id_idx" ON "lessons" ("author_id");

CREATE INDEX "lessons_is_listed_idx" ON "lessons" ("is_published");

CREATE INDEX "lessons_position_idx" ON "lessons" ("position");

CREATE UNIQUE INDEX "lesson_articles_lesson_id_unique_idx" ON "lesson_articles" ("lesson_id");

CREATE INDEX "lesson_articles_author_id_idx" ON "lesson_articles" ("author_id");

CREATE UNIQUE INDEX "lesson_videos_lesson_id_unique_idx" ON "lesson_videos" ("lesson_id");

CREATE INDEX "lesson_videos_author_id_idx" ON "lesson_videos" ("author_id");

CREATE INDEX "lesson_files_lesson_id_idx" ON "lesson_files" ("lesson_id");

CREATE INDEX "lesson_files_author_id_idx" ON "lesson_files" ("author_id");

CREATE INDEX "lesson_files_created_at_idx" ON "lesson_files" ("created_at");

CREATE UNIQUE INDEX "lesson_files_lesson_id_name_unique_idx" ON "lesson_files" ("lesson_id", "name");

CREATE UNIQUE INDEX "language_progress_user_id_language_slug_unique_idx" ON "language_progress" ("user_id", "language_slug");

CREATE INDEX "language_progress_user_id_idx" ON "language_progress" ("user_id");

CREATE INDEX "language_progress_language_slug_idx" ON "language_progress" ("language_slug");

CREATE INDEX "language_progress_viewed_at_idx" ON "language_progress" ("viewed_at");

CREATE UNIQUE INDEX "series_progress_user_id_series_slug_language_slug_unique_idx" ON "series_progress" ("user_id", "series_slug", "language_slug");

CREATE INDEX "series_progress_user_id_idx" ON "series_progress" ("user_id");

CREATE INDEX "series_progress_series_slug_idx" ON "series_progress" ("series_slug");

CREATE INDEX "series_progress_language_slug_idx" ON "series_progress" ("language_slug");

CREATE INDEX "series_progress_language_progress_id_idx" ON "series_progress" ("language_progress_id");

CREATE INDEX "series_progress_viewed_at_idx" ON "series_progress" ("viewed_at");

CREATE UNIQUE INDEX "section_progress_user_id_language_slug_series_slug_section_id_unique_idx" ON "section_progress" ("user_id", "language_slug", "series_slug", "section_id");

CREATE INDEX "section_progress_user_id_idx" ON "section_progress" ("user_id");

CREATE INDEX "section_progress_language_slug_idx" ON "section_progress" ("language_slug");

CREATE INDEX "section_progress_series_slug_idx" ON "section_progress" ("series_slug");

CREATE INDEX "section_progress_section_id_idx" ON "section_progress" ("section_id");

CREATE INDEX "section_progress_language_progress_id_idx" ON "section_progress" ("language_progress_id");

CREATE INDEX "section_progress_series_progress_id_idx" ON "section_progress" ("series_progress_id");

CREATE INDEX "section_progress_viewed_at_idx" ON "section_progress" ("viewed_at");

CREATE UNIQUE INDEX "lesson_progress_user_id_language_slug_series_slug_section_id_lesson_id_unique_idx" ON "lesson_progress" ("user_id", "language_slug", "series_slug", "section_id", "lesson_id");

CREATE INDEX "lesson_progress_user_id_idx" ON "lesson_progress" ("user_id");

CREATE INDEX "lesson_progress_language_slug_idx" ON "lesson_progress" ("language_slug");

CREATE INDEX "lesson_progress_series_slug_idx" ON "lesson_progress" ("series_slug");

CREATE INDEX "lesson_progress_section_id_idx" ON "lesson_progress" ("section_id");

CREATE INDEX "lesson_progress_lesson_id_idx" ON "lesson_progress" ("lesson_id");

CREATE INDEX "lesson_progress_language_progress_id_idx" ON "lesson_progress" ("language_progress_id");

CREATE INDEX "lesson_progress_series_progress_id_idx" ON "lesson_progress" ("series_progress_id");

CREATE INDEX "lesson_progress_section_progress_id_idx" ON "lesson_progress" ("section_progress_id");

CREATE INDEX "lesson_progress_viewed_at_idx" ON "lesson_progress" ("viewed_at");

CREATE UNIQUE INDEX "certificates_user_id_series_slug_unique_idx" ON "certificates" ("user_id", "series_slug");

CREATE INDEX "certificates_user_id_idx" ON "certificates" ("user_id");

CREATE INDEX "certificates_language_slug_idx" ON "certificates" ("language_slug");

CREATE INDEX "certificates_series_slug_idx" ON "certificates" ("series_slug");

ALTER TABLE "auth_providers" ADD FOREIGN KEY ("email") REFERENCES "users" ("email") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "languages" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series" ADD FOREIGN KEY ("language_slug") REFERENCES "languages" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_pictures" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_pictures" ADD FOREIGN KEY ("series_id") REFERENCES "series" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "sections" ADD FOREIGN KEY ("language_slug") REFERENCES "languages" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "sections" ADD FOREIGN KEY ("series_slug") REFERENCES "series" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "sections" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lessons" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lessons" ADD FOREIGN KEY ("language_slug") REFERENCES "languages" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lessons" ADD FOREIGN KEY ("series_slug") REFERENCES "series" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lessons" ADD FOREIGN KEY ("section_id") REFERENCES "sections" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_articles" ADD FOREIGN KEY ("lesson_id") REFERENCES "lessons" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_articles" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_videos" ADD FOREIGN KEY ("lesson_id") REFERENCES "lessons" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_videos" ADD FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_files" ADD FOREIGN KEY ("lesson_id") REFERENCES "lessons" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "language_progress" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "language_progress" ADD FOREIGN KEY ("language_slug") REFERENCES "languages" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_progress" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_progress" ADD FOREIGN KEY ("series_slug") REFERENCES "series" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_progress" ADD FOREIGN KEY ("language_slug") REFERENCES "languages" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "series_progress" ADD FOREIGN KEY ("language_progress_id") REFERENCES "language_progress" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "section_progress" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "section_progress" ADD FOREIGN KEY ("language_slug") REFERENCES "languages" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "section_progress" ADD FOREIGN KEY ("series_slug") REFERENCES "series" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "section_progress" ADD FOREIGN KEY ("section_id") REFERENCES "sections" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "section_progress" ADD FOREIGN KEY ("language_progress_id") REFERENCES "language_progress" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "section_progress" ADD FOREIGN KEY ("series_progress_id") REFERENCES "series_progress" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_progress" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_progress" ADD FOREIGN KEY ("language_slug") REFERENCES "languages" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_progress" ADD FOREIGN KEY ("series_slug") REFERENCES "series" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_progress" ADD FOREIGN KEY ("section_id") REFERENCES "sections" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_progress" ADD FOREIGN KEY ("lesson_id") REFERENCES "lessons" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_progress" ADD FOREIGN KEY ("language_progress_id") REFERENCES "language_progress" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_progress" ADD FOREIGN KEY ("series_progress_id") REFERENCES "series_progress" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "lesson_progress" ADD FOREIGN KEY ("section_progress_id") REFERENCES "section_progress" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "certificates" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "certificates" ADD FOREIGN KEY ("language_slug") REFERENCES "languages" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "certificates" ADD FOREIGN KEY ("series_slug") REFERENCES "series" ("slug") ON DELETE CASCADE ON UPDATE CASCADE;
