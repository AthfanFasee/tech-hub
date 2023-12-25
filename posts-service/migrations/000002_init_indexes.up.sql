CREATE INDEX IF NOT EXISTS posts_title_idx ON posts USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS posts_userid_idx ON posts(user_id);
CREATE INDEX IF NOT EXISTS comments_postid_idx ON comments(post_id);