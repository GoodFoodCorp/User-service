DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS favorites;
ALTER TABLE profiles DROP COLUMN IF EXISTS age;
ALTER TABLE profiles DROP COLUMN IF EXISTS avatar_url;
