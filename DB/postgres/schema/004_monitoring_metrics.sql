CREATE TABLE monitoring_snapshots (
    id BIGSERIAL PRIMARY KEY,
    metric_type TEXT NOT NULL, -- sessions, locks, tablespaces, sql
    target_db TEXT NOT NULL,
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    data JSONB NOT NULL
);

CREATE INDEX idx_metrics_type_time 
ON monitoring_snapshots(metric_type, collected_at DESC);
