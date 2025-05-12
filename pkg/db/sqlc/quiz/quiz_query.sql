----------------------------- questions --------------------------------
-- name: CreateQuestion :exec
INSERT INTO question (id,
    room_code,
    topic,
    question_count,
    question_data,
    time_limit,
    created_by,
    updated_by, 
    created_on, 
    updated_on)
VALUES ($7,$1, $2, $3, $4, $5,$6, $8,NOW(), NOW());

-- name: UpdateQuestionByID :exec
UPDATE question
SET 
  topic = $2,
  question_count=$5,
  question_data = $3, 
  time_limit = $6,
  updated_on = NOW(),
  updated_by = $4
WHERE id = $1;

-- name: GetQuestionsByRoomCode :one
SELECT *
FROM question
WHERE room_code = $1
ORDER BY created_on ASC;


---------------------------- answers --------------------------------------

-- name: CreateAnswer :exec
INSERT INTO answer (id,
    room_code,
    user_id,
    question_id,
    question_data_id,
    answer_option,
    is_correct,
    answer_time,
    created_by,
    updated_by,
    created_on,
    updated_on)
VALUES ($10,$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW());

-- name: UpdateAnswer :exec
UPDATE answer
SET answer_option = $2,
    is_correct = $3,
    answer_time = $4,
    updated_on = NOW(),
    updated_by = $5
WHERE id = $1;

-- name: GetAnswerByRoomCodeAndUserID :many
SELECT *
FROM answer
WHERE room_code = $1
AND user_id = $2;

-- name: ListAnswersByRoomCode :many
SELECT *
FROM answer
WHERE room_code = $1
ORDER BY created_on ASC;
