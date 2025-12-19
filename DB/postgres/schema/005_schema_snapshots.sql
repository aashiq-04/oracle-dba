CREATE TABLE oracle_schema_snapshots (
    id BIGSERIAL PRIMARY KEY,
    oracle_schema TEXT NOT NULL,
    snapshot_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    objects JSONB NOT NULL
);

CREATE INDEX idx_schema_snapshots 
ON oracle_schema_snapshots(oracle_schema, snapshot_time DESC);
