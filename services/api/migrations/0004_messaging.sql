-- OpenEdu secure parent-teacher messaging (v0.4).
-- Records are school and class scoped and never publicly exposed.
CREATE TABLE school_messages (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 school_id uuid NOT NULL,
 class_id uuid NOT NULL,
 sender_id uuid NOT NULL REFERENCES users(id),
 recipient_id uuid NOT NULL REFERENCES users(id),
 content text NOT NULL CHECK(char_length(content) BETWEEN 1 AND 2000),
 created_at timestamptz NOT NULL DEFAULT now(),
 read_at timestamptz,
 CHECK(sender_id <> recipient_id),
 FOREIGN KEY(school_id,class_id) REFERENCES classes(school_id,id)
);
CREATE INDEX school_messages_inbox ON school_messages(school_id,recipient_id,created_at DESC);
CREATE INDEX school_messages_outbox ON school_messages(school_id,sender_id,created_at DESC);
CREATE TABLE security_events (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 actor_id uuid REFERENCES users(id),
 school_id uuid REFERENCES schools(id),
 event_type text NOT NULL,
 occurred_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX security_events_scope ON security_events(school_id,occurred_at DESC);