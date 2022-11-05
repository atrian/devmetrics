CREATE TABLE IF NOT EXISTS public.metrics
(
    id VARCHAR not null,
    type VARCHAR not null,
    delta BIGINT null,
    value DOUBLE PRECISION null,
    PRIMARY KEY (id, type)
);