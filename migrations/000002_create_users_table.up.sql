CREATE TABLE IF NOT EXISTS
public.users (
    id SERIAL PRIMARY KEY,
    uid UUID NOT NULL,
    record_id INTEGER REFERENCES records(id)
);