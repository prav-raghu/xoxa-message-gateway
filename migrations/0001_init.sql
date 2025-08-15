-- Initial migration for messages table
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    to VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(50) NOT NULL
);
