-- name: GetIngredients :many
SELECT
  *
FROM
  Ingredients;

-- name: GetIngredientsByIds :many
SELECT
  *
FROM
  Ingredients
WHERE
  id = ANY(sqlc.arg('ids')::uuid[]);

-- WHERE
--   name = sqlc.arg('name');
-- date created is current date
-- name: CreateIngredient :one
INSERT INTO
  Ingredients (
    creator_id,
    name,
    unit,
    storage_loc,
    ingredient_type,
    image_path
  )
VALUES
  (
    sqlc.arg ('creator_id'),
    sqlc.arg ('name'),
    sqlc.arg ('unit'),
    sqlc.arg ('storage_loc'),
    sqlc.arg ('ingredient_type'),
    sqlc.narg ('image_path')
  )
RETURNING
  *;

-- name: SearchIngredients :many
SELECT
  *
FROM
  Ingredients
WHERE
  name ILIKE '%' || sqlc.arg('name') || '%';
  
-- name: DeleteIngredient :exec
DELETE FROM Ingredients
WHERE
  id = sqlc.arg('id');