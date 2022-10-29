CREATE TABLE IF NOT EXISTS public.metrics
(
    id VARCHAR PRIMARY KEY,
    type VARCHAR not null,
    delta BIGINT null,
    value DOUBLE PRECISION null
);