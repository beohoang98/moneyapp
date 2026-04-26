CREATE TABLE incomes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount INTEGER NOT NULL CHECK(amount > 0),
    date DATE NOT NULL,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_incomes_date ON incomes(date);
CREATE INDEX idx_incomes_category_id ON incomes(category_id);
