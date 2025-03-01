-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
  id UUID NOT NULL PRIMARY KEY,
  username TEXT NOT NULL,
  refresh_token TEXT NOT NULL,
  user_type TEXT NOT NULL, -- normal user or bot
  user_meta JSONB NOT NULL,
  premium BOOLEAN NOT NULL DEFAULT false,
  is_active BOOLEAN NOT NULL,
  is_deleted BOOLEAN NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS room (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_code TEXT NOT NULL,
  room_owner UUID NOT NULL,
  room_members JSONB NOT NULL,       
  room_chat JSONB NOT NULL,          
  leaderboard JSONB NOT NULL,        
  room_meta JSONB NOT NULL,
  room_lock BOOLEAN NOT NULL DEFAULT false,
  is_active BOOLEAN NOT NULL,
  is_deleted BOOLEAN NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS room_member (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id UUID NOT NULL,
  user_id UUID NOT NULL ,
  is_bot BOOLEAN NOT NULL DEFAULT false,
  joined_on TIMESTAMP NOT NULL,
  is_kicked BOOLEAN NOT NULL,
  is_active BOOLEAN NOT NULL,
  is_deleted BOOLEAN NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS leaderboard (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id UUID NOT NULL ,
  user_id UUID NOT NULL,
  score INT NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL,
  UNIQUE (room_id, user_id)
);

CREATE TABLE IF NOT EXISTS question (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id UUID NOT NULL,
  question_data JSONB NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS answer (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id UUID NOT NULL,
  user_id UUID NOT NULL,
  question_id UUID NOT NULL,
  answer_option INT NOT NULL,
  is_correct BOOLEAN NOT NULL,
  answer_time TIMESTAMP NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS answer;
DROP TABLE IF EXISTS question;
DROP TABLE IF EXISTS leaderboard;
DROP TABLE IF EXISTS room_member;
DROP TABLE IF EXISTS room;
DROP TABLE IF EXISTS users;



-- +goose StatementEnd
