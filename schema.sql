CREATE TABLE users (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email    TEXT NOT NULL UNIQUE,
  name     TEXT NOT NULL
);

CREATE TABLE recipes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  description TEXT,
  steps       TEXT[] NOT NULL,
  ingredients TEXT[] NOT NULL,
  creatorId   UUID REFERENCES users(id)
);

CREATE TABLE pantry_items (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name           TEXT NOT NULL,
  quantity       NUMERIC NOT NULL,
  unit           TEXT NOT NULL,
  price          NUMERIC,
  expiration_ms  BIGINT,
  category       TEXT NOT NULL,
  userId         UUID REFERENCES users(id)
);