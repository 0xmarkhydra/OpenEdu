-- Personal reading state. Progress is not shared with other readers, schools or parents.
CREATE TABLE reader_progress (
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 lesson_id uuid NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
 is_completed boolean NOT NULL DEFAULT false,
 is_bookmarked boolean NOT NULL DEFAULT false,
 last_opened_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(user_id,lesson_id)
);
CREATE INDEX reader_progress_user_recent ON reader_progress(user_id,updated_at DESC);
