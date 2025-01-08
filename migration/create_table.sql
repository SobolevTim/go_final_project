CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date INTEGER,
    title TEXT,
    comment TEXT,
    repeat TEXT
);
CREATE INDEX scheduler_date ON scheduler (date);