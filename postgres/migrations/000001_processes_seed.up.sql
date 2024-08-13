CREATE TABLE process_seeds (
    process_id TEXT PRIMARY KEY NOT NULL,
    oab TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE process_seeds IS 'This table stores the seed data for processes';
