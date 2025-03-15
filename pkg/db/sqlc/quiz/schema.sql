CREATE TABLE IF NOT EXISTS question (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_code TEXT NOT NULL,
  topic TEXT,
  question_count INT NOT NULL,
  question_data JSONB NOT NULL,
  time_limit INT NOT NULL, -- max time for each question
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL
);

-- everybody in the room's answers will be stored here
CREATE TABLE IF NOT EXISTS answer (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_code TEXT NOT NULL,
  user_id UUID NOT NULL,
  question_id UUID NOT NULL,
  question_data_id UUID NOT NULL,
  answer_option INT NOT NULL,
  is_correct BOOLEAN NOT NULL,
  answer_time TIMESTAMP NOT NULL,
  created_on TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_on TIMESTAMP NOT NULL DEFAULT NOW(),
  created_by TEXT NOT NULL,
  updated_by TEXT NOT NULL
);
