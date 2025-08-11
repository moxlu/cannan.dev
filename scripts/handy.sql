-- Using sqlite3 (Precompiled Binaries for Linux)
-- https://www.sqlite.org/download.html
--
-- To initialise:
-- cd ../run 
-- ./sqlite3 cannan.db
-- .read ../scripts/schema.sql
-- .read init.sql

-- Should use every time to enforce key integrity across tables
PRAGMA foreign_keys = 1;

-- Move weekly challenges into all challenges (Thursday evenings)
UPDATE CHALLENGES SET challenge_featured = 0;

-- Renew all invites (dev env only)
UPDATE INVITES SET invite_claimed_by = NULL, invite_claimed_time = NULL;

-- Generate starter batch of invites
WITH RECURSIVE
  cnt(x) AS (
    SELECT 1
    UNION ALL
    SELECT x+1 FROM cnt WHERE x < 50
  )
INSERT INTO INVITES (invite_token)
SELECT hex(randomblob(16)) FROM cnt;
UPDATE INVITES SET invite_expiry = '2025-09-01 10:00:00';