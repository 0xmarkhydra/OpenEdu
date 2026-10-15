-- Auth version invalidates all prior JWTs on password rotation.
ALTER TABLE users ADD COLUMN auth_version bigint NOT NULL DEFAULT 1;
