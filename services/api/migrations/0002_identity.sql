-- Authentication identities and credential throttling. No default accounts.
CREATE TABLE user_credentials (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  password_hash text NOT NULL,
  changed_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_canonical_unique ON users (lower(email));
CREATE TABLE platform_roles (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role = 'platform_admin'),
  is_active boolean NOT NULL DEFAULT true,
  PRIMARY KEY (user_id, role)
);
CREATE TABLE login_failures (
  key text PRIMARY KEY,
  failures integer NOT NULL DEFAULT 0,
  last_attempt timestamptz NOT NULL DEFAULT now(),
  blocked_until timestamptz
);
CREATE INDEX login_failures_expiry ON login_failures(last_attempt);
