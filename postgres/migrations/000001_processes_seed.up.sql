BEGIN;

CREATE TABLE process_seeds (
    process_id TEXT NOT NULL,
    oab TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (process_id, oab)
);

COMMENT ON TABLE process_seeds IS 'This table stores the seed data for processes';

COMMIT;
