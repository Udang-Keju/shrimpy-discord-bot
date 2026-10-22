-- Welcome channel greetings are now an embed (plus optional plain text) instead of a rendered card,
-- so they need the same title/description fields as the other embed-capable messages.
ALTER TABLE welcome_config ADD COLUMN embed_title TEXT;
ALTER TABLE welcome_config ADD COLUMN embed_description TEXT;
