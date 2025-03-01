-- name: CreateRoom :one
INSERT INTO room (
    id, 
    room_code, 
    room_owner, 
    room_members, 
    room_chat, 
    leaderboard, 
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
