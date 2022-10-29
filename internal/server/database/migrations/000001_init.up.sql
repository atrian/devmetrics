CREATE TABLE IF NOT EXISTS public.metrics
(
    id VARCHAR PRIMARY KEY,
    type VARCHAR not null,
    delta DOUBLE PRECISION null,
    value INT null
);