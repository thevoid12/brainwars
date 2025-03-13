-- name: GetUserDetailsByID :many
SELECT * FROM users
WHERE id = $1 AND is_deleted = false;
