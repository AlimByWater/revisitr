-- One-off data rewrite after migrating to revisitr.ru with root-level paths.
-- Rewrites persisted storage URLs: /revisitr/storage/{key} -> /storage/{key}
--
-- Run ONCE on the NEW server AFTER restoring the Postgres dump, e.g.:
--   docker exec -i infra-postgres-1 psql -U revisitr -d revisitr < rewrite-storage-urls.sql
--
-- Idempotent: replace() on already-rewritten rows is a no-op, and the WHERE
-- guards limit work to rows still containing the old prefix.

BEGIN;

UPDATE campaigns
SET media_url = replace(media_url, '/revisitr/storage/', '/storage/')
WHERE media_url LIKE '%/revisitr/storage/%';

UPDATE campaigns
SET content = replace(content::text, '/revisitr/storage/', '/storage/')::jsonb
WHERE content::text LIKE '%/revisitr/storage/%';

UPDATE emoji_items
SET image_url = replace(image_url, '/revisitr/storage/', '/storage/')
WHERE image_url LIKE '%/revisitr/storage/%';

UPDATE menu_items
SET image_url = replace(image_url, '/revisitr/storage/', '/storage/')
WHERE image_url LIKE '%/revisitr/storage/%';

-- Verification: should return 0 across the board.
SELECT 'campaigns.media_url' AS col, count(*) FROM campaigns WHERE media_url LIKE '%/revisitr/storage/%'
UNION ALL SELECT 'campaigns.content', count(*) FROM campaigns WHERE content::text LIKE '%/revisitr/storage/%'
UNION ALL SELECT 'emoji_items.image_url', count(*) FROM emoji_items WHERE image_url LIKE '%/revisitr/storage/%'
UNION ALL SELECT 'menu_items.image_url', count(*) FROM menu_items WHERE image_url LIKE '%/revisitr/storage/%';

COMMIT;
