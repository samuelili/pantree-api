CREATE TYPE UNIT_TYPE AS ENUM('count_qtr', 'volume_ml', 'mass_g');

CREATE TYPE MEASURE_TYPE AS ENUM('metric', 'imperial');

CREATE TYPE LOC_TYPE AS ENUM('pantry', 'fridge', 'freezer');

CREATE TYPE GROC_TYPE AS ENUM(
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

-- users
CREATE TABLE
  Users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    date_joined DATE NOT NULL,
    pref_measure MEASURE_TYPE NOT NULL DEFAULT 'metric'
  );

-- recipes
CREATE TABLE
  Recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    creator_id UUID REFERENCES Users (id),
    date_created DATE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    steps TEXT[] NOT NULL,
    allergens TEXT[],
    cooking_time NUMERIC NOT NULL,
    serving_size NUMERIC NOT NULL,
    image_path TEXT
  );

-- favorite recipes
CREATE TABLE
  Favorites (
    user_id UUID REFERENCES Users (id),
    recipe_id UUID REFERENCES Recipes (id),
    PRIMARY KEY (user_id, recipe_id)
  );

-- recipe -> ingredients link table
CREATE TABLE
  RecipeIngredients (
    recipe_id UUID REFERENCES Recipes (id) ON DELETE CASCADE,
    ingredient_id UUID REFERENCES Ingredients (id) ON DELETE CASCADE,
    quantity NUMERIC NOT NULL,
    author_unit_type UNIT_TYPE NOT NULL,
    author_measure_type MEASURE_TYPE NOT NULL,
    PRIMARY KEY (recipe_id, ingredient_id)
  );

-- all ingredients
CREATE TABLE
  Ingredients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    creator_id UUID REFERENCES Users (id),
    name TEXT NOT NULL,
    unit UNIT_TYPE NOT NULL,
    storage_loc LOC_TYPE NOT NULL,
    ingredient_type GROC_TYPE NOT NULL,
    image_path TEXT
  );

-- user inventories
CREATE TABLE
  UserItems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID REFERENCES Users (id) ON DELETE CASCADE,
    ingredient_id UUID REFERENCES Ingredients (id),
    quantity NUMERIC NOT NULL,
    price NUMERIC(1000, 2) CHECK (price > 0),
    expiration_date TIMESTAMP,
    last_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
  );

-- recipe ingredients view
CREATE VIEW
  RecipeIngredientsView AS
SELECT
  i.name,
  i.unit,
  i.storage_loc,
  i.ingredient_type,
  r.quantity,
  r.recipe_id
FROM
  RecipeIngredients r
  JOIN Ingredients i ON r.ingredient_id = i.id;

-- user pantry view
CREATE VIEW
  UserPantryView AS
SELECT
  u.pref_measure AS user_measurement_system,
  i.name AS ingredient_name,
  ui.quantity,
  ui.expiration_date,
  i.unit,
  i.storage_loc,
  i.ingredient_type,
  ui.last_modified
FROM
  Users u
  JOIN UserItems ui ON u.id = ui.user_id
  JOIN Ingredients i ON ui.ingredient_id = i.id;