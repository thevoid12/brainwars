----------------------------- questions --------------------------------
-- name: CreateQuestion :exec
INSERT INTO question (room_id,
    topic,
    question_count,
    question_data,
    created_by,
    updated_by, 
    created_on, 
    updated_on)
VALUES ($1, $2, $3, $4, $5,$6, NOW(), NOW());

-- name: UpdateQuestionByID :exec
UPDATE question
SET 
  topic = $2,
  question_count=$5,
  question_data = $3, 
  updated_on = NOW(),
  updated_by = $4
WHERE id = $1;

-- name: ListQuestionsByRoomID :many
SELECT *
FROM question
WHERE room_id = $1
ORDER BY created_on ASC;

---------------------------- answers --------------------------------------

-- name: CreateAnswer :exec
INSERT INTO answer (room_id,
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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW());

-- name: UpdateAnswer :exec
UPDATE answer
SET answer_option = $2,
    is_correct = $3,
    answer_time = $4,
    updated_on = NOW(),
    updated_by = $5
WHERE id = $1;

-- name: ListAnswersByRoomID :many
SELECT *
FROM answer
WHERE room_id = $1
ORDER BY created_on ASC;
