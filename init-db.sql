-- Initialize PokeBot MySQL Database
-- This file is automatically run by docker-compose on first start

USE pokebot;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_created_at (created_at),
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_user_name ON users(name);
CREATE INDEX IF NOT EXISTS idx_user_created ON users(created_at);

-- Optional: Sample data for testing
-- INSERT INTO users (name) VALUES ('Test User 1');
-- INSERT INTO users (name) VALUES ('Test User 2');

-- Set proper permissions
GRANT ALL PRIVILEGES ON pokebot.* TO 'pokebot_user'@'%';
FLUSH PRIVILEGES;
