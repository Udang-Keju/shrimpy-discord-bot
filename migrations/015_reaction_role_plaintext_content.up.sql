-- Reaction role messages can carry plain text (Discord markdown) in addition to, or instead of, an embed.
ALTER TABLE reaction_role_messages ADD COLUMN content TEXT;
