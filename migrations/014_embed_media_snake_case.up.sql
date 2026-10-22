-- migrations/014_embed_media_snake_case.up.sql
-- API JSON is snake_case everywhere. EmbedMedia is persisted as JSONB and passed straight
-- through to the dashboard, so rewrite its stored "iconUrl" keys to "icon_url".

UPDATE welcome_config
SET embed_media = replace(embed_media::text, '"iconUrl"', '"icon_url"')::jsonb
WHERE embed_media::text LIKE '%"iconUrl"%';

UPDATE ticket_panels
SET embed_media = replace(embed_media::text, '"iconUrl"', '"icon_url"')::jsonb
WHERE embed_media::text LIKE '%"iconUrl"%';

UPDATE ticket_categories
SET ticket_open_media = replace(ticket_open_media::text, '"iconUrl"', '"icon_url"')::jsonb
WHERE ticket_open_media::text LIKE '%"iconUrl"%';

UPDATE reaction_role_messages
SET embed_media = replace(embed_media::text, '"iconUrl"', '"icon_url"')::jsonb
WHERE embed_media::text LIKE '%"iconUrl"%';
