-- name: CreateRateLimit :exec
INSERT INTO rate_limit (
  id, 
  user_id,
  allowed,
  tries, 
  premium, 
  is_deleted, 
  created_on, 
  updated_on, 
  created_by, 
  updated_by    
)   
VALUES ($1, $2,$3, $4, $5, $6, $7, $8, $9,$10);

-- name: GetRateLimitByUserID :many
SELECT * FROM rate_limit
WHERE user_id = $1 AND is_deleted = false;

-- name: UpdateRateLimitTriesByUserID :exec
UPDATE rate_limit
SET 
  tries = $2,
  updated_on = NOW(),
  updated_by = $3
WHERE user_id = $1;

-- name: UpdateRateLimitByUserID :exec
UPDATE rate_limit
SET
  allowed = $2,
  tries = $3,
  premium = $4,
  is_deleted = $5,
  updated_on = NOW(),
  updated_by = $3
WHERE user_id = $1;



