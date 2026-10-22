-- migrations/014_embed_media_snake_case.down.sql

UPDATE welcome_config
SET embed_media = replace(embed_media::text, '"icon_url"', '"iconUrl"')::jsonb
WHERE embed_media::text LIKE '%"icon_url"%';

UPDATE ticket_panels
SET embed_media = replace(embed_media::text, '"icon_url"', '"iconUrl"')::jsonb
WHERE embed_media::text LIKE '%"icon_url"%';

UPDATE ticket_categories
SET ticket_open_media = replace(ticket_open_media::text, '"icon_url"', '"iconUrl"')::jsonb
WHERE ticket_open_media::text LIKE '%"icon_url"%';

UPDATE reaction_role_messages
SET embed_media = replace(embed_media::text, '"icon_url"', '"iconUrl"')::jsonb
WHERE embed_media::text LIKE '%"icon_url"%';
