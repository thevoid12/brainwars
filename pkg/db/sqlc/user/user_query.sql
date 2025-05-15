-- name: GetUserDetailsByID :many
SELECT * FROM users
WHERE id = $1 AND is_deleted = false;

-- name: GetUserDetailsByAuth0SubID :many
SELECT * FROM users
WHERE auth0_sub = $1 AND is_deleted = false;

-- name: CreateNewUser :exec
INSERT INTO users (
  id,
  auth0_sub,
  username,
  user_type,
  bot_type,
  user_meta,
  premium,
  is_active,
  is_deleted,
  created_by,
  updated_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
);



