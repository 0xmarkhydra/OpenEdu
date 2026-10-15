-- OpenEdu school operations. Additive tenant-scoped data models.
CREATE TABLE attendance_records (
 school_id uuid NOT NULL,
 class_id uuid NOT NULL,
 student_id uuid NOT NULL,
 on_date date NOT NULL,
 status text NOT NULL CHECK(status IN ('present','absent','late','excused')),
 note text NOT NULL DEFAULT '',
 recorded_by uuid NOT NULL REFERENCES users(id),
 updated_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(school_id,class_id,student_id,on_date),
 FOREIGN KEY(school_id,class_id,student_id) REFERENCES class_students(school_id,class_id,student_id)
);
CREATE INDEX attendance_class_day ON attendance_records(school_id,class_id,on_date);
CREATE TABLE assignments (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 school_id uuid NOT NULL,
 class_id uuid NOT NULL,
 title text NOT NULL,
 instructions text NOT NULL DEFAULT '',
 due_at timestamptz,
 created_by uuid NOT NULL REFERENCES users(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(school_id,id),
 FOREIGN KEY(school_id,class_id) REFERENCES classes(school_id,id)
);
CREATE INDEX assignments_class_created ON assignments(school_id,class_id,created_at DESC);
CREATE TABLE assignment_submissions (
 school_id uuid NOT NULL,
 assignment_id uuid NOT NULL,
 student_id uuid NOT NULL REFERENCES users(id),
 content text NOT NULL,
 grade numeric(5,2) CHECK(grade BETWEEN 0 AND 10),
 feedback text,
 submitted_at timestamptz NOT NULL DEFAULT now(),
 graded_at timestamptz,
 PRIMARY KEY(school_id,assignment_id,student_id),
 FOREIGN KEY(school_id,assignment_id) REFERENCES assignments(school_id,id)
);
CREATE TABLE school_announcements (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 school_id uuid NOT NULL REFERENCES schools(id),
 title text NOT NULL,
 content text NOT NULL,
 audience text NOT NULL DEFAULT 'all' CHECK(audience IN ('all','teacher','student','guardian')),
 created_by uuid NOT NULL REFERENCES users(id),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX announcements_school_new ON school_announcements(school_id,created_at DESC);
CREATE TABLE timetable_entries (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 school_id uuid NOT NULL,
 class_id uuid NOT NULL,
 weekday smallint NOT NULL CHECK(weekday BETWEEN 1 AND 7),
 period smallint NOT NULL CHECK(period BETWEEN 1 AND 15),
 subject_code text NOT NULL,
 teacher_id uuid REFERENCES users(id),
 room text NOT NULL DEFAULT '',
 UNIQUE(class_id,weekday,period),
 FOREIGN KEY(school_id,class_id) REFERENCES classes(school_id,id)
);
CREATE TABLE leave_requests (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 school_id uuid NOT NULL,
 class_id uuid NOT NULL,
 student_id uuid NOT NULL,
 requested_by uuid NOT NULL REFERENCES users(id),
 start_on date NOT NULL,
 end_on date NOT NULL,
 reason text NOT NULL,
 status text NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','approved','rejected','cancelled')),
 resolved_by uuid REFERENCES users(id),
 resolved_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 CHECK(end_on >= start_on),
 FOREIGN KEY(school_id,class_id,student_id) REFERENCES class_students(school_id,class_id,student_id)
);
CREATE INDEX leave_school_status ON leave_requests(school_id,status,created_at DESC);
