CREATE TABLE IF NOT EXISTS reading_progress (
    progress_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    target_id INT REFERENCES reading_target(target_id),
    current_page INT NOT NULL,
    last_update_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);