CREATE TABLE IF NOT EXISTS room (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_code TEXT NOT NULL,
  room_name TEXT,
  room_owner UUID NOT NULL,
  room_chat JSONB NOT NULL,          
  room_meta JSONB NOT NULL,
  room_lock BOOLEAN NOT NULL DEFAULT false,
  game_TYPE TEXT NOT NULL,
  room_status TEXT NOT NULL, -- game started,game ended,game about to start
  is_active BOOLEAN NOT NULL,
  is_deleted BOOLEAN NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS room_member (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_code TEXT NOT NULL,
  room_id TEXT NOT NULL, -- primary key of room table
  user_id UUID NOT NULL ,
  is_bot BOOLEAN NOT NULL DEFAULT false,
  joined_on TIMESTAMP NOT NULL,
  room_member_status TEXT NOT NULL,
  is_active BOOLEAN NOT NULL,
  is_deleted BOOLEAN NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS leaderboard (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_code TEXT NOT NULL ,
  user_id UUID NOT NULL,
  score FLOAT NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL,
  UNIQUE (room_code, user_id)
);
