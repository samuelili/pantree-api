-- name: GetRecipe :one
SELECT * FROM Recipes
WHERE id = sqlc.arg('id') LIMIT 1;

-- name: ListRecipes :many
SELECT * FROM Recipes;

-- date created is current date
-- name: CreateRecipe :one
INSERT INTO Recipes (
  creator_id, date_created, name, description, steps, 
  allergens, cooking_time, serving_size, favorite, image_path  
) VALUES (
  sqlc.arg('creator_id'), 
  CURRENT_DATE, 
  sqlc.arg('name'), 
  sqlc.arg('description'), 
  sqlc.arg('steps'),
  sqlc.narg('allergens'),
  sqlc.arg('cooking_time'),
  sqlc.arg('serving_size'),
  sqlc.arg('favorite'),
  sqlc.narg('image_path')
)
RETURNING *;

-- all fields are optional except for id
-- name: UpdateRecipe :exec
UPDATE Recipes
SET
  creator_id = COALESCE(sqlc.narg('creator_id'), creator_id),
  date_created = COALESCE(sqlc.narg('date_created'), date_created),
  name = COALESCE(sqlc.narg('name'), name),
  description = COALESCE(sqlc.narg('description'), description),
  steps = COALESCE(sqlc.narg('steps'), steps),
  allergens = COALESCE(sqlc.narg('allergens'), allergens),
  cooking_time = COALESCE(sqlc.narg('cooking_time'), cooking_time),
  serving_size = COALESCE(sqlc.narg('serving_size'), serving_size),
  favorite = COALESCE(sqlc.narg('favorite'), favorite),
  image_path = COALESCE(sqlc.narg('image_path'), image_path)
WHERE 
  id = sqlc.arg('id')
RETURNING *;

-- name: CreateRecipeIngredient :one
INSERT INTO RecipeIngredients (
  recipe_id, ingredient_id, quantity, author_unit_type, author_measure_type
) VALUES (
  sqlc.arg('recipe_id'),
  sqlc.arg('ingredient_id'),
  sqlc.arg('quantity'),
  sqlc.arg('author_unit_type'),
  sqlc.arg('author_measure_type')
)
RETURNING *;

-- $1: recipe_id
-- name: GetRecipeIngredients :many
SELECT 
  name, 
  unit, 
  storage_loc, 
  ingredient_type,
  quantity,
  recipe_id
FROM
  RecipeIngredientsView
WHERE
  recipe_id = sqlc.arg('recipe_id');

-- name: CreateUser :one
INSERT INTO Users (
  email, name, date_joined, pref_measure
) VALUES (
  sqlc.arg('email'),
  sqlc.arg('name'),
  CURRENT_DATE,
  sqlc.arg('pref_measure')
)
RETURNING *;

-- name: UpdateUser :exec
UPDATE Users
SET
  email = COALESCE(sqlc.narg('email'), email),
  name = COALESCE(sqlc.narg('name'), name),
  date_joined = COALESCE(sqlc.narg('date_joined'), date_joined),
  pref_measure = COALESCE(sqlc.narg('pref_measure')::measure_type, pref_measure)
WHERE 
  id = sqlc.arg('id')
RETURNING *;

-- select by either id or email
-- name: GetUser :one
SELECT 
  * 
FROM 
  Users u
WHERE
  (sqlc.narg('id')::uuid IS NOT NULL AND u.id = sqlc.narg('id')::uuid) 
  OR (sqlc.narg('email')::text IS NOT NULL AND u.email = sqlc.narg('email')::text);

-- select by either id or email
-- name: GetUserPantry :many
SELECT
  user_measurement_system,
  ingredient_name,
  quantity,
  expiration_date,
  unit, 
  storage_loc,
  ingredient_type
FROM
  UserPantryView
WHERE
  (sqlc.narg('user_id')::uuid IS NOT NULL AND id = sqlc.narg('user_id')::uuid) 
  OR (sqlc.narg('email')::text IS NOT NULL AND email = sqlc.narg('email')::text);