TRUNCATE TABLE bookmarks CASCADE;

ALTER TABLE bookmarks ADD COLUMN user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);