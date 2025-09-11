-- $1: recipeId
-- name: GetRecipe :one
SELECT * FROM recipes
WHERE id = $1 LIMIT 1;

-- $1: userId
-- name: ListRecipes :many
SELECT * FROM recipes;

-- $1: name
-- $2: description
-- $3: steps
-- $4: ingredients
-- $5: creatorId
-- name: CreateRecipe :one
INSERT INTO recipes (
  name, description, steps, ingredients, creatorId
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- all fields are optional except for id
-- name: UpdateRecipe :exec
UPDATE recipes
SET
  name = COALESCE(sqlc.narg('name'), name),
  description = COALESCE(sqlc.narg('description'), description),
  steps = COALESCE(sqlc.narg('steps'), steps),
  ingredients = COALESCE(sqlc.narg('ingredients'), ingredients)
WHERE id = sqlc.arg('id')
RETURNING *;