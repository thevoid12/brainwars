CREATE TABLE IF NOT EXISTS users (
  id UUID NOT NULL PRIMARY KEY,
  auth0_sub TEXT, -- unique customer id
  username TEXT NOT NULL,
  user_type TEXT NOT NULL, -- normal user or bot
  bot_type TEXT, -- if a user is a bot he will have a bot type
  user_meta JSONB NOT NULL,
  premium BOOLEAN NOT NULL DEFAULT false,
  is_active BOOLEAN NOT NULL,
  is_deleted BOOLEAN NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL
);