-- Align default vision model with project standard (qwen3-vl:4b).
-- Only updates rows still on the legacy default so customized model names are preserved.
UPDATE scanning_settings
SET model = 'qwen3-vl:4b', updated_at = CURRENT_TIMESTAMP
WHERE id = 1 AND model = 'moondream:1.8b';
