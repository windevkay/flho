CREATE INDEX IF NOT EXISTS workflows_name_idx ON workflows USING GIN (to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS workflows_states_idx ON workflows USING GIN (states);