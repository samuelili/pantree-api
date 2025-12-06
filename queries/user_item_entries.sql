-- name: GetUserItemEntryIngredientIdsForUser :many
SELECT
  ingredient_id
FROM
  UserItemEntries
WHERE
  user_id = sqlc.arg('user_id');