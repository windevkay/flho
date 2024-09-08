CREATE TABLE IF NOT EXISTS states (
    id bigserial PRIMARY KEY,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	updated_at timestamp(0) with time zone,
    deleted_at timestamp(0) with time zone,
    workflow_id bigint NOT NULL REFERENCES workflows ON DELETE CASCADE,
    name text NOT NULL,
    retryUrl text,
    retryAfter integer
);