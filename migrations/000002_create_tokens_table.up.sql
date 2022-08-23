CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES auth_user ON DELETE CASCADE,
    expiry TIMESTAMP(0) WITH TIME ZONE NOT  NULL,
    scope TEXT NOT NULL
);