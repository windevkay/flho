CREATE TABLE IF NOT EXISTS identities (
    id bigserial PRIMARY KEY,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	updated_at timestamp(0) with time zone,
	deleted_at timestamp(0) with time zone,
	uuid text NOT NULL,
	name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,      
	activated boolean NOT NULL,
	version integer NOT NULL DEFAULT 1 
);