----------------------------------- room table ---------------------------------------------------------------------
-- name: CreateRoom :one
INSERT INTO room (
  id, 
  room_code,
  room_name, 
  room_owner, 
  room_chat, 
  room_meta, 
  room_lock, 
  is_active, 
  is_deleted, 
  created_on, 
  updated_on, 
  created_by, 
  updated_by,
  game_type,
  room_status
) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW(), $10, $11, $12,$13)
RETURNING *;

-- name: ListRoomByUserID :many
SELECT * FROM room
WHERE room_owner = $1 AND is_deleted = false;

-- name: GetRoomByID :many
SELECT * FROM room
WHERE id = $1 AND is_deleted = false;

-- name: GetRoomByRoomCode :many
SELECT * FROM room
WHERE room_code = $1 AND is_deleted = false;

-- name: UpdateRoomByID :exec
UPDATE room
SET 
  room_name = $2,
  room_chat = $3,
  room_meta = $4,
  room_lock = $5,
  is_active = $6,
  updated_on = NOW(),
  updated_by = $7,
  game_type = $8,
  room_status = $9
WHERE id = $1;

-- name: UpdateRoomMetaByRoomCode :exec
UPDATE room
SET 
  room_meta = $2,
  updated_on = NOW(),
  updated_by = $3
WHERE room_code = $1 AND is_deleted = false;
-------------------------------------- Room Member ------------------------------------------------------------------------

-- name: CreateRoomMember :one
INSERT INTO room_member (
  id, 
  room_code,
  room_id,
  user_id, 
  is_bot, 
  joined_on, 
  room_member_status, 
  is_active, 
  is_deleted, 
  created_on, 
  updated_on, 
  created_by, 
  updated_by    
)   
VALUES ($1, $2,$10, $3, $4, NOW(), $5, $6, $7, NOW(), NOW(), $8, $9)
RETURNING *;

-- name: ListRoomMembersByRoomCode :many
SELECT * FROM room_member INNER JOIN users ON room_member.user_id = users.id
WHERE room_code = $1 AND room_member.is_deleted = false;

-- name: GetRoomMemberByRoomCodeAndUserID :many
SELECT * FROM room_member INNER JOIN users ON room_member.user_id = users.id
WHERE room_code = $1 AND user_id = $2 AND room_member.is_deleted = false;

-- name: GetRoomMemberByID :many
SELECT * FROM room_member INNER JOIN users ON room_member.user_id = users.id
WHERE room_member.id = $1 AND room_member.is_deleted = false;

-- name: UpdateRoomMemberByID :exec
UPDATE room_member
SET 
  room_member_status = $2,
  is_active = $3,
  updated_on = NOW(),
  updated_by = $4
WHERE id = $1;

-- name: UpdateRoomMemberByRoomCodeAndUserID :exec
UPDATE room_member
SET 
  room_member_status = $2,
  is_active = $3,
  updated_on = NOW(),
  updated_by = $4
WHERE room_code = $1 AND user_id = $5 AND is_deleted=false;

--------------------------------------- leaderboard ------------------------------------------------------------------------
-- name: CreatLeaderBoard :exec
INSERT INTO leaderboard (
  id, 
  room_code,
  user_id, 
  score, 
  created_on, 
  updated_on, 
  created_by, 
  updated_by,
  is_deleted    
) 
VALUES ($1, $2, $3, $4, NOW(), NOW(), $5, $6,$7);

-- name: ListLeaderBoardByRoomCode :many
SELECT * FROM leaderboard
WHERE room_code = $1 AND is_deleted = false 
ORDER BY score DESC;

-- name: GetLeaderBoardByID :many
SELECT * FROM leaderboard
WHERE id = $1 AND is_deleted = false;

-- name: UpdateLeaderBoardScoreByID :exec
UPDATE leaderboard
SET 
  score = $2,
  updated_on = NOW(),
  updated_by = $3
WHERE id = $1 AND is_deleted = false;

-- name: UpdateLeaderBoardScoreByUserIDAndRoomCode :exec
UPDATE leaderboard
SET 
  score = $3,
  updated_on = NOW(),
  updated_by = $4
WHERE room_code = $1 AND user_id = $2 AND is_deleted = false;
