CREATE TABLE IF NOT EXISTS metrics
(
    id                  INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    metric_name TEXT    NOT NULL UNIQUE,
    metric_type TEXT    NOT NULL,
    metric_value        FLOAT,
    metric_delta        INTEGER,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
