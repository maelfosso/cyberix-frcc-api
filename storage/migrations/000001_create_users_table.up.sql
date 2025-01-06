CREATE TABLE IF NOT EXISTS users (
  id INTEGER Primary Key Generated Always as Identity,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  quality TEXT NOT NULL,
  phone TEXT UNIQUE NOT NULL,
  organization TEXT NOT NULL,

  created_at TIMESTAMPZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPZ NOT NULL DEFAULT NOW()
)

