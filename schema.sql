CREATE TYPE UNIT_TYPE AS ENUM ('count_qtr', 'volume_ml', 'mass_g');
CREATE TYPE MEASURE_TYPE AS ENUM ('metric', 'imperial');
CREATE TYPE LOC_TYPE AS ENUM ('pantry', 'fridge', 'freezer');
CREATE TYPE GROC_TYPE AS ENUM (
  'meat/seafood', 
  'produce', 
  'dairy/eggs', 
  'prepared', 
  'essentials', 
  'bakery', 
  'snacks', 
  'frozen', 
  'beverages', 
  'desserts',
  'alcohol'
);

CREATE TABLE Users (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email          TEXT UNIQUE NOT NULL,
  name           TEXT NOT NULL,
  date_joined    DATE NOT NULL,
  pref_measure   MEASURE_TYPE NOT NULL
);

CREATE TABLE Recipes (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  creator_id   UUID REFERENCES Users(id),
  date_created DATE NOT NULL,
  name         TEXT NOT NULL,
  description  TEXT,
  steps        TEXT[] NOT NULL
);

-- recipe -> ingredients link table
CREATE TABLE RecipeIngredients (
  recipe_id       UUID REFERENCES Recipes(id),
  ingredient_id   UUID REFERENCES Ingredients(id),
  quantity        NUMERIC NOT NULL,
  PRIMARY KEY (recipe_id, ingredient_id)
);

-- all ingredients
CREATE TABLE Ingredients (
  id UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  name             TEXT NOT NULL,
  unit             UNIT_TYPE NOT NULL,
  storage_loc      LOC_TYPE NOT NULL,
  ingredient_type  GROC_TYPE NOT NULL
);

-- user inventories
CREATE TABLE UserItems ( 
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          UUID REFERENCES Users(id) ON DELETE CASCADE,
  ingredient_id    UUID REFERENCES Ingredients(id),
  quantity         NUMERIC NOT NULL,
  price            NUMERIC(1000, 2) CHECK (price > 0),
  expiration_date  DATE
);