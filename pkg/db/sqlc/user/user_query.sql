-- -- name: ListTest :many
-- SELECT * FROM test
-- ORDER BY name;

-- -- name: UpdateTest :exec
-- UPDATE test
--   set name = $2,
--   bio = $3
-- WHERE id = $1;

-- -- name: UpdateAndReturnTest :one
-- UPDATE test
--   set name = $2,
--   bio = $3
-- WHERE id = $1
-- RETURNING *;

-- -- name: CreateTest :one
-- INSERT INTO test (
--   name, bio
-- ) VALUES (
--   $1, $2
-- )
-- RETURNING *;
