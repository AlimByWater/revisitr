-- +goose Up

-- Add JSONB content column for composite messages to campaigns
ALTER TABLE campaigns ADD COLUMN content JSONB;

-- Add content to campaign_templates
ALTER TABLE campaign_templates ADD COLUMN content JSONB;

-- Add content to campaign_variants
ALTER TABLE campaign_variants ADD COLUMN content JSONB;

-- Add content to auto_scenarios
ALTER TABLE auto_scenarios ADD COLUMN content JSONB;

-- Migrate existing text campaigns to new format
-- Detect media type from file extension to avoid data corruption
UPDATE campaigns
SET content = jsonb_build_object(
    'parts', jsonb_build_array(
        CASE
            WHEN media_url IS NOT NULL AND media_url != '' THEN
                jsonb_build_object(
                    'type', CASE
                        WHEN media_url ~* '\.(mp4|mov|avi|webm)$' THEN 'video'
                        WHEN media_url ~* '\.(gif)$' THEN 'animation'
                        WHEN media_url ~* '\.(pdf|doc|docx|xls|xlsx|csv|txt|zip|rar)$' THEN 'document'
                        WHEN media_url ~* '\.(mp3|ogg|wav|flac|aac)$' THEN 'audio'
                        WHEN media_url ~* '\.(webp)$' THEN 'sticker'
                        ELSE 'photo'
                    END,
                    'text', message,
                    'media_url', media_url,
                    'parse_mode', 'Markdown'
                )
            ELSE
                jsonb_build_object(
                    'type', 'text',
                    'text', message,
                    'parse_mode', 'Markdown'
                )
        END
    ),
    'buttons', COALESCE(buttons, '[]'::jsonb)
)
WHERE content IS NULL AND message != '';

-- Migrate campaign_templates
UPDATE campaign_templates
SET content = jsonb_build_object(
    'parts', jsonb_build_array(
        CASE
            WHEN media_url IS NOT NULL AND media_url != '' THEN
                jsonb_build_object(
                    'type', CASE
                        WHEN media_url ~* '\.(mp4|mov|avi|webm)$' THEN 'video'
                        WHEN media_url ~* '\.(gif)$' THEN 'animation'
                        WHEN media_url ~* '\.(pdf|doc|docx|xls|xlsx|csv|txt|zip|rar)$' THEN 'document'
                        WHEN media_url ~* '\.(mp3|ogg|wav|flac|aac)$' THEN 'audio'
                        WHEN media_url ~* '\.(webp)$' THEN 'sticker'
                        ELSE 'photo'
                    END,
                    'text', message, 'media_url', media_url, 'parse_mode', 'Markdown')
            ELSE
                jsonb_build_object('type', 'text', 'text', message, 'parse_mode', 'Markdown')
        END
    ),
    'buttons', COALESCE(buttons, '[]'::jsonb)
)
WHERE content IS NULL AND message != '';

-- Migrate campaign_variants (fix: preserve existing buttons)
UPDATE campaign_variants
SET content = jsonb_build_object(
    'parts', jsonb_build_array(
        CASE
            WHEN media_url IS NOT NULL AND media_url != '' THEN
                jsonb_build_object(
                    'type', CASE
                        WHEN media_url ~* '\.(mp4|mov|avi|webm)$' THEN 'video'
                        WHEN media_url ~* '\.(gif)$' THEN 'animation'
                        WHEN media_url ~* '\.(pdf|doc|docx|xls|xlsx|csv|txt|zip|rar)$' THEN 'document'
                        WHEN media_url ~* '\.(mp3|ogg|wav|flac|aac)$' THEN 'audio'
                        WHEN media_url ~* '\.(webp)$' THEN 'sticker'
                        ELSE 'photo'
                    END,
                    'text', message, 'media_url', media_url, 'parse_mode', 'Markdown')
            ELSE
                jsonb_build_object('type', 'text', 'text', message, 'parse_mode', 'Markdown')
        END
    ),
    'buttons', COALESCE(buttons, '[]'::jsonb)
)
WHERE content IS NULL AND message != '';

-- Migrate auto_scenarios
UPDATE auto_scenarios
SET content = jsonb_build_object(
    'parts', jsonb_build_array(
        jsonb_build_object('type', 'text', 'text', message, 'parse_mode', 'Markdown')
    )
)
WHERE content IS NULL AND message != '';

-- +goose Down
ALTER TABLE campaigns DROP COLUMN IF EXISTS content;
ALTER TABLE campaign_templates DROP COLUMN IF EXISTS content;
ALTER TABLE campaign_variants DROP COLUMN IF EXISTS content;
ALTER TABLE auto_scenarios DROP COLUMN IF EXISTS content;
