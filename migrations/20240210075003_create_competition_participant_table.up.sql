CREATE TABLE IF NOT EXISTS competition_participant (
    participant_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    competition_id INT REFERENCES competition(competition_id),
    total_pages INT NOT NULL
);