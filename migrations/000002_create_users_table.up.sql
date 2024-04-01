CREATE TABLE IF NOT EXISTS
public.users (
    uid UUID PRIMARY KEY,
    record_id INTEGER REFERENCES records(id)
);