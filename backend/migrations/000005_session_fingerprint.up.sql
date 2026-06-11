ALTER TABLE sessions ADD COLUMN ip_address VARCHAR(255);
ALTER TABLE sessions ADD COLUMN user_agent TEXT;
ALTER TABLE sessions ADD COLUMN last_activity_at TIMESTAMP;
