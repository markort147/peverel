-- CREATE TABLE IF NOT EXISTS groups (
--   id     INTEGER PRIMARY KEY AUTOINCREMENT,
--   name   TEXT NOT NULL UNIQUE
-- );

CREATE TABLE IF NOT EXISTS tasks (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  name           TEXT NOT NULL,
  description    TEXT NOT NULL,
  period         INTEGER NOT NULL,              -- days
  last_completed TEXT NOT NULL                 -- RFC3339 UTC
  -- group_id       INTEGER REFERENCES groups(id)  -- nullable
);
