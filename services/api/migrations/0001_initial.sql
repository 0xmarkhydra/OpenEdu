-- FreeEducation initial schema. Append new migrations; never edit after production deployment.
CREATE TABLE schools (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  slug text UNIQUE NOT NULL,
  name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text UNIQUE NOT NULL,
  display_name text NOT NULL,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','locked')),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE school_memberships (
  school_id uuid NOT NULL REFERENCES schools(id),
  user_id uuid NOT NULL REFERENCES users(id),
  role text NOT NULL CHECK (role IN ('school_admin','principal','teacher','student','guardian','content_admin','content_reviewer')),
  is_active boolean NOT NULL DEFAULT true,
  PRIMARY KEY (school_id,user_id,role)
);
CREATE TABLE academic_years (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id uuid NOT NULL REFERENCES schools(id),
  name text NOT NULL,
  starts_on date NOT NULL,
  ends_on date NOT NULL,
  UNIQUE (school_id,name),
  UNIQUE (school_id,id),
  CHECK (ends_on > starts_on)
);
CREATE TABLE classes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id uuid NOT NULL REFERENCES schools(id),
  academic_year_id uuid NOT NULL REFERENCES academic_years(id),
  name text NOT NULL,
  grade smallint NOT NULL CHECK (grade BETWEEN 1 AND 12),
  UNIQUE (school_id,academic_year_id,name),
  UNIQUE (school_id,id),
  FOREIGN KEY (school_id,academic_year_id) REFERENCES academic_years(school_id,id)
);
CREATE TABLE class_students (
  school_id uuid NOT NULL,
  class_id uuid NOT NULL,
  student_id uuid NOT NULL REFERENCES users(id),
  student_role text NOT NULL DEFAULT 'student' CHECK (student_role='student'),
  enrolled_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (school_id,class_id,student_id),
  FOREIGN KEY (school_id,class_id) REFERENCES classes(school_id,id),
  FOREIGN KEY (school_id,student_id,student_role) REFERENCES school_memberships(school_id,user_id,role)
);
CREATE TABLE teacher_assignments (
  school_id uuid NOT NULL,
  class_id uuid NOT NULL,
  teacher_id uuid NOT NULL REFERENCES users(id),
  teacher_role text NOT NULL DEFAULT 'teacher' CHECK (teacher_role='teacher'),
  subject_code text NOT NULL,
  PRIMARY KEY (school_id,class_id,teacher_id,subject_code),
  FOREIGN KEY (school_id,class_id) REFERENCES classes(school_id,id),
  FOREIGN KEY (school_id,teacher_id,teacher_role) REFERENCES school_memberships(school_id,user_id,role)
);
CREATE TABLE guardian_links (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id uuid NOT NULL REFERENCES schools(id),
  guardian_id uuid NOT NULL REFERENCES users(id),
  student_id uuid NOT NULL REFERENCES users(id),
  guardian_role text NOT NULL DEFAULT 'guardian' CHECK (guardian_role='guardian'),
  student_role text NOT NULL DEFAULT 'student' CHECK (student_role='student'),
  verified_at timestamptz,
  revoked_at timestamptz,
  UNIQUE (school_id,guardian_id,student_id),
  FOREIGN KEY (school_id,guardian_id,guardian_role) REFERENCES school_memberships(school_id,user_id,role),
  FOREIGN KEY (school_id,student_id,student_role) REFERENCES school_memberships(school_id,user_id,role),
  CHECK (guardian_id <> student_id)
);
CREATE INDEX idx_guardian_links_guardian ON guardian_links(school_id,guardian_id) WHERE verified_at IS NOT NULL AND revoked_at IS NULL;
CREATE TABLE books (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id uuid REFERENCES schools(id),
  title text NOT NULL,
  grade smallint NOT NULL CHECK (grade BETWEEN 1 AND 12),
  subject_code text NOT NULL,
  rights_owner text NOT NULL,
  rights_license text NOT NULL,
  rights_source text NOT NULL,
  visibility text NOT NULL CHECK (visibility IN ('public','school')),
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','review','approved','published','archived')),
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  published_at timestamptz,
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK ((visibility = 'public' AND school_id IS NULL) OR (visibility = 'school' AND school_id IS NOT NULL)),
  CHECK (char_length(trim(rights_owner)) > 0 AND char_length(trim(rights_license)) > 0 AND char_length(trim(rights_source)) > 0)
);
CREATE INDEX idx_books_published ON books(grade,subject_code,school_id) WHERE status='published';
CREATE INDEX idx_books_school ON books(school_id,status,updated_at DESC);
CREATE TABLE chapters (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  book_id uuid NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  title text NOT NULL,
  position integer NOT NULL CHECK (position > 0),
  UNIQUE (book_id,position)
);
CREATE TABLE lessons (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  chapter_id uuid NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
  title text NOT NULL,
  position integer NOT NULL CHECK (position > 0),
  content_type text NOT NULL DEFAULT 'html' CHECK (content_type IN ('html','text','pdf')),
  content_body text,
  object_key text,
  UNIQUE (chapter_id,position),
  CHECK ((content_type = 'pdf' AND object_key IS NOT NULL) OR (content_type IN ('text','html') AND content_body IS NOT NULL))
);
CREATE TABLE book_audit_log (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  book_id uuid NOT NULL REFERENCES books(id),
  actor_id uuid NOT NULL REFERENCES users(id),
  from_status text NOT NULL,
  to_status text NOT NULL,
  at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_book_audit_book ON book_audit_log(book_id,at DESC);
