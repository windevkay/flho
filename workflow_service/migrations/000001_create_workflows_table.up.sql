CREATE TABLE IF NOT EXISTS workflows (
    id bigserial PRIMARY KEY,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	updated_at timestamp(0) with time zone,
	identity_id bigint NOT NULL REFERENCES identities,
	uniqueId text NOT NULL,
	name text NOT NULL,         
	active boolean NOT NULL,
	version integer NOT NULL DEFAULT 1 
);