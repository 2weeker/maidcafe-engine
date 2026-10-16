ALTER TABLE users ADD COLUMN art TEXT NOT NULL DEFAULT '' CHECK (char_length(art) <= 2100);
