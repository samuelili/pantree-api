CREATE TYPE MEASURE_SYS ENUM('metric', 'imperial');

CREATE TABLE users (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email    TEXT NOT NULL UNIQUE,
  name     TEXT NOT NULL,
  measurement MEASURE_SYS NOT NULL
);

CREATE TABLE recipes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  description TEXT,
  steps       TEXT[] NOT NULL,
  ingredients TEXT[] NOT NULL,
  creatorId   UUID REFERENCES users(id)
);

CREATE TABLE recipe_ingredients (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  recipe_id UUID REFERENCES recipes(id),
  

)

CREATE TABLE pantry_items (
  id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  item_name              TEXT NOT NULL,
  expiration_date        DATE,
  purchase_date          DATE,
  quantity               NUMERIC NOT NULL,
  unit                   TEXT NOT NULL,
  price                  NUMERIC,
  storage_category       TEXT NOT NULL,
  userId                 UUID REFERENCES users(id)
);