-- OpenEdu server-graded single-choice exams: answer keys never leave the staff API.
CREATE TABLE quizzes (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 school_id uuid NOT NULL,
 class_id uuid NOT NULL,
 title text NOT NULL CHECK(char_length(title) BETWEEN 3 AND 180),
 duration_minutes smallint NOT NULL CHECK(duration_minutes BETWEEN 5 AND 180),
 due_at timestamptz NOT NULL,
 status text NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','published')),
 created_by uuid NOT NULL REFERENCES users(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(school_id,id),
 FOREIGN KEY(school_id,class_id) REFERENCES classes(school_id,id)
);
CREATE INDEX quizzes_class_status ON quizzes(school_id,class_id,status,due_at);
CREATE TABLE quiz_questions (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 quiz_id uuid NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
 position smallint NOT NULL CHECK(position BETWEEN 1 AND 100),
 prompt text NOT NULL CHECK(char_length(prompt) BETWEEN 3 AND 2000),
 option_a text NOT NULL CHECK(char_length(option_a) BETWEEN 1 AND 500),
 option_b text NOT NULL CHECK(char_length(option_b) BETWEEN 1 AND 500),
 option_c text NOT NULL CHECK(char_length(option_c) BETWEEN 1 AND 500),
 option_d text NOT NULL CHECK(char_length(option_d) BETWEEN 1 AND 500),
 correct_option smallint NOT NULL CHECK(correct_option BETWEEN 0 AND 3),
 UNIQUE(quiz_id,position)
);
CREATE TABLE quiz_attempts (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 school_id uuid NOT NULL,
 quiz_id uuid NOT NULL,
 student_id uuid NOT NULL REFERENCES users(id),
 started_at timestamptz NOT NULL DEFAULT now(),
 expires_at timestamptz NOT NULL,
 submitted_at timestamptz,
 answers jsonb,
 grade numeric(5,2) CHECK(grade BETWEEN 0 AND 10),
 UNIQUE(quiz_id,student_id),
 FOREIGN KEY(school_id,quiz_id) REFERENCES quizzes(school_id,id)
);
CREATE INDEX quiz_attempts_school_student ON quiz_attempts(school_id,student_id,started_at DESC);
