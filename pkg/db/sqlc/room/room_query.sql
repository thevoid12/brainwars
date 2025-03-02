-- name: CreateRoom :one
INSERT INTO room (
    id, 
    room_code,
    room_name, 
    room_owner, 
    room_members, 
    room_chat, 
    room_meta, 
    room_lock, 
    is_active, 
    is_deleted, 
    created_on, 
    updated_on, 
    created_by, 
    updated_by    
) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW(), $11, $12)
RETURNING *;

-- name: ListRoomByUserID :many
SELECT * FROM room
WHERE room_owner = $1 AND is_deleted = false;

-- name: GetRoomByID :many
SELECT * FROM room
WHERE id = $1 AND is_deleted = false;

-- name: UpdateRoomByID :exec
UPDATE room
SET 
    room_name = $2,
    room_members = $3,
    room_chat = $4,
    room_meta = $5,
    room_lock = $6,
    is_active = $7,
    updated_on = NOW(),
    updated_by = $8
WHERE id = $1;