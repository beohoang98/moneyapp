CREATE TABLE IF NOT EXISTS scanning_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    enabled INTEGER NOT NULL DEFAULT 0,
    base_url TEXT NOT NULL DEFAULT 'http://localhost:11434/v1',
    model TEXT NOT NULL DEFAULT 'qwen3-vl:4b',
    api_key TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT OR IGNORE INTO scanning_settings (id) VALUES (1);
