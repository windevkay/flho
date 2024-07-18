CREATE TABLE IF NOT EXISTS workflows (
    id bigserial PRIMARY KEY,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	updated_at timestamp(0) with time zone,
	uniqueId text NOT NULL,
	name text NOT NULL,
	states text[] NOT NULL,
	startstate text NOT NULL,  
	endstate text NOT NULL,
	retrywebhook text,        
	retryAfter integer,          
	active boolean NOT NULL,
	version integer NOT NULL DEFAULT 1 
);