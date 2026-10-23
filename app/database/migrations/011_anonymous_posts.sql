UPDATE posts SET deleted_at = created_at WHERE user_id IS NULL AND deleted_at IS NULL;
