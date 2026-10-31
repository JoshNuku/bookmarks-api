DROP INDEX IF EXISTS idx_bookmarks_user_id;

ALTER TABLE bookmarks 
DROP COLUMN IF EXISTS user_id;