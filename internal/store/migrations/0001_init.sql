CREATE TABLE users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  email         TEXT NOT NULL UNIQUE,
  name          TEXT NOT NULL DEFAULT '',
  password_hash TEXT NOT NULL DEFAULT '',
  role          TEXT NOT NULL DEFAULT 'member',
  created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at    TEXT
);

CREATE TABLE sessions (
  id         TEXT PRIMARY KEY,
  user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE api_tokens (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name         TEXT NOT NULL DEFAULT '',
  token_hash   TEXT NOT NULL UNIQUE,
  prefix       TEXT NOT NULL DEFAULT '',
  last_used_at TEXT,
  created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

-- Box-wide settings (title, setup flag); everything else is personal.
CREATE TABLE settings (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL DEFAULT ''
);

CREATE TABLE user_settings (
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  key     TEXT NOT NULL,
  value   TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (user_id, key)
);

CREATE TABLE cache (
  key        TEXT PRIMARY KEY,
  value      TEXT NOT NULL DEFAULT '',
  fetched_at TEXT NOT NULL
);


-- A person whose chart you keep: you (is_me, one per account), a partner,
-- family, friends. Birth date and time are local to the birth place (tz);
-- an unknown time means no houses or rising sign.
CREATE TABLE people (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  relation    TEXT NOT NULL DEFAULT 'friend',  -- me | partner | family | friend | colleague | other
  is_me       INTEGER NOT NULL DEFAULT 0,
  birth_date  TEXT NOT NULL,                   -- YYYY-MM-DD
  birth_time  TEXT NOT NULL DEFAULT '',        -- HH:MM, '' = unknown
  place       TEXT NOT NULL DEFAULT '',
  country     TEXT NOT NULL DEFAULT '',
  lat         REAL NOT NULL DEFAULT 0,
  lon         REAL NOT NULL DEFAULT 0,
  tz          TEXT NOT NULL DEFAULT '',        -- IANA zone of the birth place
  color       TEXT NOT NULL DEFAULT '',
  notes       TEXT NOT NULL DEFAULT '',
  favorite    INTEGER NOT NULL DEFAULT 0,
  source      TEXT NOT NULL DEFAULT 'manual',  -- manual | contacts | import | <pocket app slug>
  external_id TEXT NOT NULL DEFAULT '',
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX people_user_name ON people(user_id, name);
CREATE UNIQUE INDEX people_user_me ON people(user_id) WHERE is_me = 1;
CREATE UNIQUE INDEX people_external ON people(user_id, source, external_id) WHERE external_id <> '';

-- One check-in a day: how the day felt, with the sky of that day kept next
-- to it (so the insights compare mood with the Moon).
CREATE TABLE journal (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  date        TEXT NOT NULL,                   -- local YYYY-MM-DD
  mood        INTEGER NOT NULL DEFAULT 0,      -- 1..5, 0 = not set
  energy      INTEGER NOT NULL DEFAULT 0,      -- 1..5, 0 = not set
  note        TEXT NOT NULL DEFAULT '',
  tags        TEXT NOT NULL DEFAULT '',
  sky         TEXT NOT NULL DEFAULT '{}',      -- JSON JournalSky
  source      TEXT NOT NULL DEFAULT 'app',     -- app | import | <pocket app slug>
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE UNIQUE INDEX journal_user_date ON journal(user_id, date);

-- Written readings (AI or composed) a person kept: a day, a chart, a match,
-- the year ahead, or an answer to a question.
CREATE TABLE readings (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  person_id  INTEGER REFERENCES people(id) ON DELETE CASCADE,
  other_id   INTEGER REFERENCES people(id) ON DELETE CASCADE,
  kind       TEXT NOT NULL,                    -- daily | natal | match | year | question
  ref        TEXT NOT NULL DEFAULT '',         -- the date, year or question
  title      TEXT NOT NULL DEFAULT '',
  body       TEXT NOT NULL DEFAULT '',         -- markdown
  model      TEXT NOT NULL DEFAULT '',
  saved      INTEGER NOT NULL DEFAULT 0,       -- kept by the person (else a cache)
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX readings_user_kind ON readings(user_id, kind, created_at);
CREATE INDEX readings_user_person ON readings(user_id, person_id);
