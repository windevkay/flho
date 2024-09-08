CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    identity_id bigint NOT NULL REFERENCES identities,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL
);