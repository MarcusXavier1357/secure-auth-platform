ALTER TABLE sessions DROP COLUMN IF EXISTS last_activity_at;
ALTER TABLE sessions DROP COLUMN IF EXISTS user_agent;
ALTER TABLE sessions DROP COLUMN IF EXISTS ip_address;
