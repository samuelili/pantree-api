-- name: GetRecipe :one
SELECT * FROM Recipes
WHERE id = sqlc.arg('id') LIMIT 1;

-- name: ListRecipes :many
SELECT * FROM Recipes;

-- date created is current date
-- name: CreateRecipe :one
INSERT INTO Recipes (
  creator_id, date_created, name, description, steps    
) VALUES (
  sqlc.arg('creator_id'), 
  CURRENT_DATE, 
  sqlc.arg('name'), 
  sqlc.arg('description'), 
  sqlc.arg('steps')
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
  steps = COALESCE(sqlc.narg('steps'), steps)
WHERE 
  id = sqlc.arg('id')
RETURNING *;

-- $1: recipe_id
-- name: GetRecipeIngredients :many
SELECT 
  i.name, 
  i.unit, 
  i.storage_loc, 
  i.ingredient_type,
  r.quantity
FROM
  RecipeIngredients r
JOIN
  Ingredients i
ON
  r.ingredient_id = i.id
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
  pref_measure = COALESCE(sqlc.narg('pref_measure'), pref_measure)
WHERE 
  id = sqlc.arg('id')
RETURNING *;

-- select by either id or email
-- name: GetUser :one
SELECT 
  * 
FROM 
  Users
WHERE
  (sqlc.narg('id')::uuid IS NOT NULL AND u.id = sqlc.narg('id')::uuid) 
  OR (sqlc.narg('email')::text IS NOT NULL AND u.email = sqlc.narg('email')::text);

-- select by either id or email
-- name: GetUserPantry :many
SELECT
  -- names, quantities, exp date, unit, storage location, ingredient type
  u.pref_measure AS user_measurement_system,
  i.name AS ingredient_name,
  ui.quantity,
  ui.expiration_date,
  i.unit,
  i.storage_loc,
  i.ingredient_type
FROM
  Users u
JOIN
  UserItems ui
ON
  u.id = ui.user_id
JOIN
  Ingredients i
ON
  ui.ingredient_id = i.id
WHERE
  (sqlc.narg('user_id')::uuid IS NOT NULL AND id = sqlc.narg('user_id')::uuid) 
  OR (sqlc.narg('email')::text IS NOT NULL AND email = sqlc.narg('email')::text);