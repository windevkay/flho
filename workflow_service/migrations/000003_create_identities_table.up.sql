CREATE TABLE IF NOT EXISTS workflow_identity_identities (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    uuid text NOT NULL
);