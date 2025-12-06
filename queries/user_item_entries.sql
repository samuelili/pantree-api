-- name: GetUserItemEntryIdsForUser :many
SELECT
  id
FROM
  UserItemEntries
WHERE
  user_id = sqlc.arg('user_id');