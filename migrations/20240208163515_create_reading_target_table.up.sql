CREATE TABLE IF NOT EXISTS reading_target (
    target_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    target_pages_per_interval INT NOT NULL
);