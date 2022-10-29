CREATE TABLE IF NOT EXISTS public.metrics
(
    id VARCHAR PRIMARY KEY,
    type VARCHAR not null,
    delta INT null,
    value DOUBLE PRECISION null
);