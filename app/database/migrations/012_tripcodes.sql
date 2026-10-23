ALTER TABLE posts
    ADD COLUMN name TEXT NOT NULL DEFAULT '' CHECK (char_length(name) <= 30),
    ADD COLUMN trip TEXT NOT NULL DEFAULT '' CHECK (char_length(trip) <= 10);
