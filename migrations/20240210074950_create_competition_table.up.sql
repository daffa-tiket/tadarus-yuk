CREATE TABLE IF NOT EXISTS competition (
    competition_id SERIAL PRIMARY KEY,
    competition_name VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL
);